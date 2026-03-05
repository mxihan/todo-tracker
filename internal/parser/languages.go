// Package parser 定义多语言注释语法支持
package parser

// LanguageConfig 语言注释配置
type LanguageConfig struct {
	Extensions     []string // 文件扩展名
	SingleLine     []string // 单行注释标记
	MultiLineStart string   // 多行注释开始
	MultiLineEnd   string   // 多行注释结束
	Name           string   // 语言名称
}

// DefaultLanguages 返回默认语言配置
func DefaultLanguages() map[string]*LanguageConfig {
	return map[string]*LanguageConfig{
		// Go
		".go": {
			Extensions:     []string{".go"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "Go",
		},

		// JavaScript/TypeScript
		".js": {
			Extensions:     []string{".js"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "JavaScript",
		},
		".ts": {
			Extensions:     []string{".ts"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "TypeScript",
		},
		".jsx": {
			Extensions:     []string{".jsx"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "JSX",
		},
		".tsx": {
			Extensions:     []string{".tsx"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "TSX",
		},

		// Python
		".py": {
			Extensions: []string{".py"},
			SingleLine: []string{"#", "# "},
			Name:       "Python",
		},
		".pyw": {
			Extensions: []string{".pyw"},
			SingleLine: []string{"#"},
			Name:       "Python",
		},

		// Java
		".java": {
			Extensions:     []string{".java"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "Java",
		},
		".kt": {
			Extensions:     []string{".kt"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "Kotlin",
		},
		".scala": {
			Extensions:     []string{".scala"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "Scala",
		},

		// C/C++
		".c": {
			Extensions:     []string{".c"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "C",
		},
		".cpp": {
			Extensions:     []string{".cpp", ".cc", ".cxx"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "C++",
		},
		".h": {
			Extensions:     []string{".h", ".hpp"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "C/C++ Header",
		},

		// C#
		".cs": {
			Extensions:     []string{".cs"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "C#",
		},

		// Ruby
		".rb": {
			Extensions: []string{".rb"},
			SingleLine: []string{"#"},
			Name:       "Ruby",
		},
		".rake": {
			Extensions: []string{".rake"},
			SingleLine: []string{"#"},
			Name:       "Ruby",
		},

		// Rust
		".rs": {
			Extensions:     []string{".rs"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "Rust",
		},

		// PHP
		".php": {
			Extensions:     []string{".php"},
			SingleLine:     []string{"//", "#"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "PHP",
		},

		// Swift
		".swift": {
			Extensions:     []string{".swift"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "Swift",
		},

		// Objective-C
		".m": {
			Extensions:     []string{".m"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "Objective-C",
		},

		// Shell
		".sh": {
			Extensions: []string{".sh"},
			SingleLine: []string{"#"},
			Name:       "Shell",
		},
		".bash": {
			Extensions: []string{".bash"},
			SingleLine: []string{"#"},
			Name:       "Bash",
		},
		".zsh": {
			Extensions: []string{".zsh"},
			SingleLine: []string{"#"},
			Name:       "Zsh",
		},

		// PowerShell
		".ps1": {
			Extensions: []string{".ps1"},
			SingleLine: []string{"#"},
			Name:       "PowerShell",
		},

		// Perl
		".pl": {
			Extensions: []string{".pl"},
			SingleLine: []string{"#"},
			Name:       "Perl",
		},
		".pm": {
			Extensions: []string{".pm"},
			SingleLine: []string{"#"},
			Name:       "Perl Module",
		},

		// Lua
		".lua": {
			Extensions:     []string{".lua"},
			SingleLine:     []string{"--"},
			MultiLineStart: "--[[",
			MultiLineEnd:   "]]",
			Name:           "Lua",
		},

		// SQL
		".sql": {
			Extensions:     []string{".sql"},
			SingleLine:     []string{"--"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "SQL",
		},

		// R
		".r": {
			Extensions: []string{".r", ".R"},
			SingleLine: []string{"#"},
			Name:       "R",
		},

		// MATLAB (.m 扩展名与 Objective-C 冲突，使用 .mat)
		".mat": {
			Extensions: []string{".mat"},
			SingleLine: []string{"%"},
			Name:       "MATLAB",
		},

		// HTML/XML
		".html": {
			Extensions:     []string{".html", ".htm"},
			MultiLineStart: "<!--",
			MultiLineEnd:   "-->",
			Name:           "HTML",
		},
		".xml": {
			Extensions:     []string{".xml"},
			MultiLineStart: "<!--",
			MultiLineEnd:   "-->",
			Name:           "XML",
		},
		".svg": {
			Extensions:     []string{".svg"},
			MultiLineStart: "<!--",
			MultiLineEnd:   "-->",
			Name:           "SVG",
		},

		// CSS/SCSS/LESS
		".css": {
			Extensions:     []string{".css"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "CSS",
		},
		".scss": {
			Extensions:     []string{".scss"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "SCSS",
		},
		".less": {
			Extensions:     []string{".less"},
			SingleLine:     []string{"//"},
			MultiLineStart: "/*",
			MultiLineEnd:   "*/",
			Name:           "LESS",
		},

		// YAML
		".yaml": {
			Extensions: []string{".yaml", ".yml"},
			SingleLine: []string{"#"},
			Name:       "YAML",
		},

		// TOML
		".toml": {
			Extensions: []string{".toml"},
			SingleLine: []string{"#"},
			Name:       "TOML",
		},

		// INI
		".ini": {
			Extensions: []string{".ini"},
			SingleLine: []string{"#", ";"},
			Name:       "INI",
		},

		// Dockerfile
		"Dockerfile": {
			Extensions: []string{"Dockerfile"},
			SingleLine: []string{"#"},
			Name:       "Dockerfile",
		},

		// Makefile
		"Makefile": {
			Extensions: []string{"Makefile", "makefile"},
			SingleLine: []string{"#"},
			Name:       "Makefile",
		},

		// Vim
		".vim": {
			Extensions: []string{".vim"},
			SingleLine: []string{"\""},
			Name:       "Vim",
		},

		// Emacs Lisp
		".el": {
			Extensions: []string{".el"},
			SingleLine: []string{";"},
			Name:       "Emacs Lisp",
		},

		// Clojure
		".clj": {
			Extensions: []string{".clj", ".cljs", ".cljc"},
			SingleLine: []string{";"},
			Name:       "Clojure",
		},

		// Haskell
		".hs": {
			Extensions:     []string{".hs"},
			SingleLine:     []string{"--"},
			MultiLineStart: "{-",
			MultiLineEnd:   "-}",
			Name:           "Haskell",
		},

		// Elixir
		".ex": {
			Extensions: []string{".ex", ".exs"},
			SingleLine: []string{"#"},
			Name:       "Elixir",
		},

		// Erlang
		".erl": {
			Extensions: []string{".erl"},
			SingleLine: []string{"%"},
			Name:       "Erlang",
		},

		// F#
		".fs": {
			Extensions:     []string{".fs", ".fsi", ".fsx"},
			SingleLine:     []string{"//"},
			MultiLineStart: "(*",
			MultiLineEnd:   "*)",
			Name:           "F#",
		},

		// Visual Basic
		".vb": {
			Extensions: []string{".vb"},
			SingleLine: []string{"'"},
			Name:       "Visual Basic",
		},

		// Vue
		".vue": {
			Extensions:     []string{".vue"},
			SingleLine:     []string{"//", "#"},
			MultiLineStart: "<!--",
			MultiLineEnd:   "-->",
			Name:           "Vue",
		},

		// Svelte
		".svelte": {
			Extensions:     []string{".svelte"},
			SingleLine:     []string{"//"},
			MultiLineStart: "<!--",
			MultiLineEnd:   "-->",
			Name:           "Svelte",
		},

		// GraphQL
		".graphql": {
			Extensions: []string{".graphql", ".gql"},
			SingleLine: []string{"#"},
			Name:       "GraphQL",
		},

		// Docker compose
		".dockerignore": {
			Extensions: []string{".dockerignore"},
			SingleLine: []string{"#"},
			Name:       "Docker Ignore",
		},

		// Git ignore
		".gitignore": {
			Extensions: []string{".gitignore"},
			SingleLine: []string{"#"},
			Name:       "Git Ignore",
		},

		// ENV files
		".env": {
			Extensions: []string{".env"},
			SingleLine: []string{"#"},
			Name:       "Environment",
		},

		// Jinja/Twig templates
		".jinja": {
			Extensions:     []string{".jinja", ".j2"},
			MultiLineStart: "{#",
			MultiLineEnd:   "#}",
			Name:           "Jinja",
		},

		// Mustache/Handlebars
		".mustache": {
			Extensions:     []string{".mustache", ".hbs"},
			MultiLineStart: "{{!",
			MultiLineEnd:   "}}",
			Name:           "Mustache",
		},
	}
}

// GetLanguageByExtension 根据扩展名获取语言配置
func GetLanguageByExtension(ext string) *LanguageConfig {
	languages := DefaultLanguages()

	// 处理特殊情况：无扩展名的文件
	switch ext {
	case "Dockerfile", "Makefile":
		return languages[ext]
	}

	// 标准化为小写
	normalized := ext
	if ext[0] != '.' {
		normalized = "." + ext
	}

	if lang, ok := languages[normalized]; ok {
		return lang
	}

	return nil
}

// IsSupported 检查文件扩展名是否支持
func IsSupported(ext string) bool {
	return GetLanguageByExtension(ext) != nil
}

// GetSupportedExtensions 获取所有支持的文件扩展名
func GetSupportedExtensions() []string {
	languages := DefaultLanguages()
	extensions := make([]string, 0, len(languages))

	for ext := range languages {
		extensions = append(extensions, ext)
	}

	return extensions
}