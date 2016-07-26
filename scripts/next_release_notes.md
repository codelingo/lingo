# Pre-release

This release comes with known issues and limitations. It should not been seen as a tool ready to point at a large repo. The purpose of the release is to: give a common baseline that we can improve on week by week; gather early feedback on DX, workflow, bugs etc; and give better external visibility to progress.

Please see the README.md for getting started. Please raise any feedback, bugs, feature requests etc as an issue on this repository.

# Functionality

* Built-in PHP facts
* Review named files or all files in PWD
* Multiple Tenets per .lingo file
* Open file/line from issue

# Known Issues:
	## Bugs:
		* Can only use with one git user.
		* Errors if swp files found in working tree.
		* Only files within pwd or named files can be reviewed (intentional restriction until streaming is sorted)
		* Horrible Error messages (due to repeated attempts)
		* Missing file errors
		* Extra tabs breaks CLQL

	## Nice To Haves:
		* Deps should be checked e.g. git version is installed.
		* Lingo init should streamline adding the git remote, prompt user for remote name and username.
		* OAuth signin on codelingo.io should be synced to git server (e.g. codelingo.io:3030) - one signin.
