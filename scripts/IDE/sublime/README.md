LingoComplete
=========

LingoComplete is a plugin for [SublimeText](http://www.sublimetext.com/) that enables highlighting and dynamic autocomplete for [CodeLingo](https://github.com/codelingo/lingo) .lingo files.

Install
-------

Install [CodeLingo](https://github.com/codelingo/lingo) and make sure the binary is on your path, as per the instructions.

Copy this directory to `~/.config/sublime-text-3/Packages` into a folder called Lingo or run `scripts/install_sublime_plugin.sh`.

If you are a developer, go to Preferences > Package Settings > Lingo and add the preference `{"codelingo_env":"dev"}`

You may need to restart sublimetext.

Reset Completions
-------

You may wish to get new autocomplete data from the CodeLingo platform, in which case you need to delete the Lingo/lexicons/<owner>/<name> file from your plugin.
