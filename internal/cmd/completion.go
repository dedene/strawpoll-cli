package cmd

import (
	"fmt"
	"os"
)

// CompletionCmd generates shell completion scripts.
type CompletionCmd struct {
	Bash CompletionBashCmd `cmd:"" help:"Generate bash completions"`
	Zsh  CompletionZshCmd  `cmd:"" help:"Generate zsh completions"`
	Fish CompletionFishCmd `cmd:"" help:"Generate fish completions"`
}

// CompletionBashCmd generates bash completion script.
type CompletionBashCmd struct{}

// Run prints the bash completion script.
func (c *CompletionBashCmd) Run() error {
	script := `_strawpoll_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local commands="version auth config poll completion"

    if [ $COMP_CWORD -eq 1 ]; then
        COMPREPLY=($(compgen -W "$commands" -- "$cur"))
    fi
}

complete -F _strawpoll_completions strawpoll
`
	fmt.Fprint(os.Stdout, script)

	return nil
}

// CompletionZshCmd generates zsh completion script.
type CompletionZshCmd struct{}

// Run prints the zsh completion script.
func (c *CompletionZshCmd) Run() error {
	script := `#compdef strawpoll

_strawpoll() {
    local -a commands
    commands=(
        'version:Show version information'
        'auth:Manage API key'
        'config:Manage configuration'
        'poll:Poll commands'
        'completion:Generate shell completions'
    )

    _arguments \
        '1: :->command' \
        '*::arg:->args'

    case $state in
        command)
            _describe 'command' commands
            ;;
    esac
}

compdef _strawpoll strawpoll
`
	fmt.Fprint(os.Stdout, script)

	return nil
}

// CompletionFishCmd generates fish completion script.
type CompletionFishCmd struct{}

// Run prints the fish completion script.
func (c *CompletionFishCmd) Run() error {
	script := `complete -c strawpoll -f

complete -c strawpoll -n '__fish_use_subcommand' -a 'version' -d 'Show version information'
complete -c strawpoll -n '__fish_use_subcommand' -a 'auth' -d 'Manage API key'
complete -c strawpoll -n '__fish_use_subcommand' -a 'config' -d 'Manage configuration'
complete -c strawpoll -n '__fish_use_subcommand' -a 'poll' -d 'Poll commands'
complete -c strawpoll -n '__fish_use_subcommand' -a 'completion' -d 'Generate shell completions'
`
	fmt.Fprint(os.Stdout, script)

	return nil
}
