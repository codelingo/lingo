import sublime
import sublime_plugin
import subprocess
import os
import json
import yaml
import re
from pprint import pprint


homePath = os.environ['HOME']
packagePath =  homePath + "/.config/sublime-text-3/Packages/lingo"

class Lingo(sublime_plugin.EventListener):
	def on_query_completions(self, view, prefix, location):
		os.environ["CODELINGO_ENV"] = "dev"
		# Will need to refactor once scopes are cleaned up
		if view.match_selector(location[0], "CLQL.lingo"):
			#TODO figure out current branch name
			#TODO get lexicon dynamically
			data = getData(view)
			completions = []
			#TODO(BlakeMScurr) leaves have no completion
			currline = view.substr(view.line(location[0]))
			line = currline
			found = ""
			m = re.search('([a-zA-Z0-9.]+):', line)
			if m:
				found = m.group(1)

			if found not in data:
				found = "match"

			for key in data:
				if key == found:
					compStub = data[key]

			if found == "match":
				compStub = list(data.keys())

			for value in compStub:
				if len(data[value]) == 0:
					branchProp = "property"
				else:
					branchProp = "branch"
				completions.append([value + "\t" + branchProp,"\n\t" + value + ": "])
			# print(completions)
			print("second last line of code")
			# TODO(BlakeMScurr) check completions append behaviour
			return (completions,sublime.INHIBIT_WORD_COMPLETIONS)
		else:
			print("Not in branch scope")
			return

def getData(view):
	maxLexicons = 50
	data = {}
	for x in range(maxLexicons):
		point = view.text_point(x, 0)
		scopes = view.scope_name(point)

		if "tenets.lingo" in scopes:
			break
		else:
			line = view.substr(view.line(point))
			m = re.search('^- ([a-zA-Z]+/[a-zA-Z.]+)(?: as ([a-zA-Z_]+))?$', line)
			if m:
				data = append_completions(data, m)				
	return data

def append_completions(data, m):
	found = m.group(1)
	if m.group(2) == "_":
		namespace = ""
	elif m.group(2) == None:
		 namespace = os.path.split(found)[1] + "."
	else:
		namespace = m.group(2) + "."

	facts = getJSONFacts(found)
	# TODO(BlakeMScurr) include logic for different namespaces having similar lexicons
	for fact in facts:
		children = []
		for child in facts[fact]:
			children.append(namespace + child)
		data[namespace + fact] = children

	return data

def getJSONFacts(lexicon):
	getDataFromPlatform = False
	fname = packagePath + '/lexicons/' + lexicon + ".json"
	ensure_dir(packagePath + "/lexicons/" + os.path.split(lexicon)[0])
	if not os.path.isfile(fname):
		subprocess.check_output([homepath + "/work/bin/lingo","list-facts", lexicon, "-f", "json", "-o", fname])	
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

def addToCompletions(completionData, contents):
	completionData['completions'].insert(len(completionData), dict(trigger=contents, contents=contents))

def ensure_dir(path):
    if not path or os.path.exists(path):
        return []
    # TODO(BlakeMScurr) use python os library which has permissions issues
    subprocess.call(["mkdir", path])

def bytes_to_json(byte):
	return json.loads(byte.decode("utf-8"))