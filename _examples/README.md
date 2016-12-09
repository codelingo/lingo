Types of queries that should be possible:

Is function / global var name unique across all packages? Handle prototypes and definitions correctly.
Find all instances of packages which are only imported once in the project
Find main() funcs which don't include a usage information block on nil args
Does a line or portion of code have a high moss score / was it copy pasted from the Internet?
Find and disallow friend classes?
Find internal private or protected variable being passed by reference out of a public function
Example Joins:
left join incl: for all functions in file, lookup documentation for each function if the function exists in the doco and set that doco to a param otherwise set the param to null but still match and return ast node
Left join excl: for all classes in module lookup doco for each class but only return classes without doco
Full outer join incl: return all classes in version A of a module and all classes in version B of a module and a combined class entity for duplicate classes
Full outer join excl: return all classes in version A and all classes in version B where those classes don't appear in both versions
Inner join: select all protected variables that are used in this child class and this other child class
Other Use cases:
count parameters in a format string and compare to vaargs -- needed for printf, prepared statements
	The parser should understand languages in languages such as SQL
	Tenets should be able to enforce some database rules like complaining about delete statements etc.
	Tenets should check/warn SQL limit clauses etc.
find pairs of modules which call functions / classes in eachother excessively ( I.e several Std deviations away from norm )
Find functions that contain absolutely no comments
Find lines of code which are being inadvertently reverted to semantically similar old versions (bug reintroduction)
Find sets of parameters which are being over-passed (passed a lot) and intelligently suggest a struct or global variable based on use
Find instantiation statements within a package which are unusual e.g. instantiate a base class instead of a child class
fuzzy semantic equivalence
exact semantic equivalence
upgrades
https://medium.com/modern-user-interfaces/architectural-tips-for-people-still-writing-ui-in-blaze-a034d7bc6d0#.j27goo6ef
Embedded systems with restricted APIs
Extending package boundaries and API contracts
If a todo is removed, check the todo was addressed.

Juju queries:
indicate there's a relationship between certupdater and apiserver
indicate the envworkermanager makes lots of those things
indicate dependencies on env or agent config

errorsf, sprintf without any args

If PR is very long, commit message body should not be empty

concise example vs PMP, pystylecheck etc
one tenet, different languages

create scripts from queires:
- scripts:
	name: "collect todos"
	bash: `exeIngestTodo $s`
	match:
	  comment:
	    text: /\/\/ TODO\(waigani\).*/ 
	    text: s

use nlp lexicon to ensure git comments are written in present tense and
formatted correctly.


- Major, Minor, Point releases: detect if public API has changed in a backwards incompatible way.

- When updating a dependency, make sure glide.yaml / godeps version is updated.

- update changelog with each significant commit

* finding who has done most work in area of codebase a direct them to handle the technical debt in that area.

- string search inside node e.g. 

func:
  line:
	content: /"github\.com.*/


little bash script to scrape out facts:

cat `find . | grep '\.lingo'` | tr '\t' ' ' | sed 's/ -//g' | grep -oE '^ *[a-z0-9\._\-]+:' | grep -oE '[a-z0-9\._\-]+:' | sort | uniq

check lifecycles new -> dying -> dead
golang
check a chan is not closed twice

defining var outside of block, but is only ever used inside block.

func in which nothing but nil is returned, yet func sig defines something should be returned.

"for _, i := range issues" 
	- don't use "i" as var name, reserved for interator int
	- don't user single letter variables
	- The singular of issues is issue or iss

naming:
 - overused package names
 - ensuring leaving hints of type in name: e.g. "namec" for chans

 make(chan error, n) - n is a code smell

 len(x) != 0 should be len(x) > 0

 any comment that has a "?" should start with a "TODO"