package util

import "github.com/codegangsta/cli"

type flagName struct {
	Long  string
	Short string
}

func (f *flagName) LongArg() string {
	return "--" + f.Long
}

func (f *flagName) ShortArg() string {
	return "-" + f.Short
}

var (

	// global flags
	TenetCfgFlg = flagName{
		"tenet-config",
		"c",
	}
	OutputTypeFlg = flagName{
		"output-type",
		"t",
	}
	OutputFlg = flagName{
		"output",
		"o",
	}
	RepoURLFlg = flagName{
		"repo-url",
		"u",
	}
	RepoPathFlg = flagName{
		"start-dir",
		"s",
	}
	OutputTemplateFlg = flagName{
		"output-template",
		"e",
	}
	LingoHomeFlg = flagName{
		"lingo-home",
		"l",
	}
	DiffFlg = flagName{
		"diff",
		"d",
	}

	//local flags
	AllFlg = flagName{
		"all",
		"a",
	}
	UpdateFlg = flagName{
		"update",
		"u",
	}
	TagsFlg = flagName{
		"tags",
		"g",
	}
	RegistryFlg = flagName{
		"registry",
		"r",
	}
	LingoFile = flagName{
		"lingo-file",
		"lf",
	}
	FormatFlg = flagName{
		"format",
		"f",
	}
	VersionFlg = flagName{
		"version",
		"v",
	}
	InteractiveFlg = flagName{
		"interactive",
		"i",
	}
	DirectoryFlg = flagName{
		"directory",
		"w",
	}
)

func (f *flagName) String() string {
	return f.Long + ", " + f.Short
}

var GlobalOptions = []cli.Flag{
	cli.StringFlag{
		Name:   RepoPathFlg.String(),
		Value:  ".",
		Usage:  "the directory to operate in, defaults to current directory",
		EnvVar: "LINGO_REPO_PATH",
	},

	cli.StringFlag{
		Name:   LingoHomeFlg.String(),
		Value:  MustLingoHome(),
		Usage:  "a directory of files needed for Lingo to operate e.g. logs and binary tenets are stored here",
		EnvVar: "LINGO_HOME",
	},

	// cli.StringFlag{
	// 	Name:   outputTemplateFlg.String(),
	// 	Value:  "",
	// 	Usage:  "a template for the output format",
	// 	EnvVar: "LINGO_OUTPUT_TEMPLATE",
	// },
	// cli.StringFlag{
	// 	Name:   repoURLFlg.String(),
	// 	Value:  "",
	// 	Usage:  "remote repository URL, if not supplied a local repository will be looked for",
	// 	EnvVar: "LINGO_REPO_URL",
	// },
	// cli.StringFlag{
	// 	Name:   outputFlg.String(),
	// 	Value:  "cli",
	// 	Usage:  "filepath to write output to. By default output will be printed to the CLI",
	// 	EnvVar: "LINGO_OUTPUT",
	// },
}
