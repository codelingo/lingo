import sublime, sublime_plugin
import os, json, re
import subprocess
from .Edit import Edit as Edit

homepath = os.environ['HOME']
packagePath =  homepath + "/.config/sublime-text-3/Packages/Lingo"

class Lingo(sublime_plugin.EventListener):
	def on_query_completions(self, view, prefix, location):
		vs = view.settings()
		setting = get_setting('codelingo_env')
		if setting:
			os.environ['CODELINGO_ENV'] = setting

		path = get_setting('path')
		if path and path not in os.environ['PATH']:
			os.environ['PATH'] += ':'
			os.environ['PATH'] += path

		lexicons = get_lexicons()
		completions = []
		# TODO (BlakeMScurr) invalidate the cache once a day, forcing a refresh on next call to plugin

		if view.match_selector(location[0], "source.lingo") and not view.match_selector(location[0], "tenets.lingo"):
			for lex in lexicons:
				completions.append([lex,lex])

			# TODO(BlakeMScurr) put tenets and lexicons completions in static file and
			# do not sublime.INHIBIT_EXPLICIT_COMPLETIONS
			completions.append(["lexicons","lexicons:"])
			completions.append(["tenets","tenets:"])
			#TODO(BlakeMScurr) use INHIBIT_WORD_COMPLETIONS on static file
			return (completions,sublime.INHIBIT_WORD_COMPLETIONS)

		# Will need to refactor once scopes are cleaned up
		if view.match_selector(location[0], "CLQL.lingo"):
			vs.set("word_separators", vs.get("word_separators").replace(".", ""))
			#TODO figure out current branch name
			# Write completions for lexicon section using "lexicons" data
			# make full python struct
			data = get_data(view)
			#TODO(BlakeMScurr) leaves have no completion
			currline = view.substr(view.line(location[0]))
			lineNumber = view.rowcol(location[0])[0]
			found = ""
			spaces = currline.count(" ")
			foundSpaces = 0
			m = None

			while spaces - 2 != foundSpaces and lineNumber >= 0:
				lineNumber -= 1
				iterating_line = view.substr(view.line(view.text_point(lineNumber, 0)))
				m = re.search('(\s*)([a-zA-Z0-9-._]+):', iterating_line)
				if m is None:
					continue
				if m.group(2) == "match":
					break
				foundSpaces = m.group(1).count(" ") // 2 * 2

			if m:
				found = m.group(2)
			if found not in data:
				found = "match"

			for key in data:
				if key == found:
					compStub = data[key]

			if found == "match":
				compStub = list(data.keys())

			for value in compStub:
				if len(data[value]) == 0:
					completions.append([value + "\t" + "property", value + ": ${1:\"property_value\"}"])
				else:
					completions.append([value + "\t" + "fact", value + ": "])
			# TODO(BlakeMScurr) check completions append behaviour
			return (completions, sublime.INHIBIT_WORD_COMPLETIONS)

	def on_pre_save(self,view):
		prev = -1
		for x in range(100):
			point = view.text_point(x, 0)
			scopes = view.scope_name(point)
			if "CLQL.lingo" in scopes:
				region = view.line(point)
				line = view.substr(region)
				m = re.search('(\s*)([a-zA-Z0-9-._]+):', line)
				if m and m.group(1).count(" ") % 2 == 1:
					with Edit(view) as edit:
						edit.insert(point, ' ')
			if prev != -1 and view.rowcol(point)[0] == view.rowcol(prev)[0]:
				break
			prev = point

	def on_text_command(self, view, command_name, args):
		if command_name == "insert" and args['characters'] == "\n":
			# TODO(BlakeMScurr) support multiple cursors
			# for pos in self.view.sel():
			#	self.view.insert(edit, pos.begin(), timestamp)

			point = view.sel()[0].begin()
			scopes = view.scope_name(point)

			previous_line = view.substr(view.line(point))
			addStr = ""
			if "tenets.lingo" in scopes:
				if re.search('  - ', previous_line) is not None:
						addStr += "  "
				if "CLQL.lingo" in scopes:
					addStr += "  "
				if re.search('tenets:', previous_line) is not None:
					addStr = "  - "
			# if we are in higher scope
			elif "source.lingo" in scopes:
				if re.search('  - ', previous_line) is not None:
					addStr += "- "
				else:
					addStr += "  - "
			if previous_line == "  - ":
				return ("run_macro_file",{"file": "res://Packages/Default/Delete to Hard BOL.sublime-macro"})
			if addStr == "":
				return None
			else:
				args['characters'] += addStr
				return (command_name, args)

def get_setting(key, default=None):
	""" Returns the setting in the following hierarchy: project setting, user setting,
	default setting.  If none are set the 'default' value passed in is returned.
	"""
	val = None
	try:
	   val = sublime.active_window().active_view().settings().get('Lingo', {}).get(key)
	except AttributeError:
		pass
	if not val:
		val = sublime.load_settings("LingoUser.sublime-settings").get(key)
	if not val:
		val = sublime.load_settings("Lingo.sublime-settings").get(key)
	if not val:
		val = default
	return val

def get_lexicons():
	# TODO(BlakeMScurr) occassionally check whether lexicon list has changed
	fname = packagePath + "/lexicons/list.json"
	if not os.path.isfile(fname):
		subprocess.check_output(["lingo","list-lexicons", "-f", "json", "-o", fname])
	with open(fname, 'r') as infile:
		data = json.load(infile)
		infile.close()
	return data

def get_data(view):
	maxLexicons = 50
	data = {}
	data = append_completions(data, "- codelingo/common as _")
	for x in range(maxLexicons):
		point = view.text_point(x, 0)
		scopes = view.scope_name(point)

		if "tenets.lingo" in scopes:
			break
		else:
			line = view.substr(view.line(point))
			data = append_completions(data, line)
	return data

def append_completions(data, line):
	m = re.search('^\s*- ([a-zA-Z]+/[a-zA-Z.]+)(?: as ([a-zA-Z_]+))?\s*$', line)
	if m:
		found = m.group(1)
		if m.group(2) == "_":
			namespace = ""
		elif m.group(2) == None:
			 namespace = os.path.split(found)[1] + "."
		else:
			namespace = m.group(2) + "."

		facts = get_json_facts(found)
		# TODO(BlakeMScurr) include logic for different namespaces having similar lexicons
		for fact in facts:
			children = []
			for child in facts[fact]:
				children.append(namespace + child)
			data[namespace + fact] = children
	return data

def get_json_facts(lexicon):
	get_dataFromPlatform = False
	fname = packagePath + '/lexicons/' + lexicon + ".json"
	ensure_dir(packagePath + "/lexicons/" + os.path.split(lexicon)[0])
	if not os.path.isfile(fname):
		subprocess.check_output(["lingo","list-facts", lexicon, "-f", "json", "-o", fname])
	with open(fname, 'r') as infile:
		data = json.load(infile)
		infile.close()

	return data

def countTabs(line):
	x = 0 
	for char in line:
		if char == "\t":
			x+=1
	return x

def ensure_dir(path):
	if not path or os.path.exists(path):
		return []
	# TODO(BlakeMScurr) use python os library which has permissions issues
	subprocess.call(["mkdir", path])

def bytes_to_json(byte):
	return json.loads(byte.decode("utf-8"))
