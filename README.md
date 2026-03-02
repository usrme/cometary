# Cometary

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/usrme/cometary)
[![Go Report Card](https://goreportcard.com/badge/github.com/usrme/cometary)](https://goreportcard.com/report/github.com/usrme/cometary)

An alternative to [Comet](https://github.com/liamg/comet) with additional features.

![Cometary - animated GIF demo](examples/demo.gif)

The way I've changed the original is for it to look and feel more like [Commitizen](https://github.com/commitizen-tools/commitizen) when invoking its sub-command `commit`. My only gripe was that the start-up speed was a tad on the slow side sometimes, given that it is Python, and that customizing the prompts wasn't as straight-forward as with Comet.

What I missed with Comet though was that Commitizen's `commit` by default keeps the values given for previous prompts on the screen, as seen in the [demo](https://github.com/commitizen-tools/commitizen/raw/master/docs/images/demo.gif), and that in and of itself was a major sticking point in continuing to use Comet.

Other minor changes include a fix to the prompt that asks for a commit message body that was misaligned and a check prior to running that confirms whether there are even any files that can be committed (i.e. are in the staging area). More improvements have been made in terms of customizing the character input limits for the scope, message, or setting a total one in general and having a visible character count for all limit types.

## Installation

- using `go install`:

```bash
go install github.com/usrme/cometary/v2@latest
```

- download a binary from the [releases](https://github.com/usrme/cometary/releases) page

- build it yourself (requires Go 1.24+):

```bash
git clone https://github.com/usrme/cometary.git
cd cometary
go build
```

## Removal

```bash
rm -f "${GOPATH}/bin/cometary"
rm -rf "${GOPATH}/pkg/mod/github.com/usrme/cometary*"
```

## Usage

There is an additional `comet.json` file that includes the prefixes and descriptions that I most prefer myself, which can be added to either the root of a repository, to one's home directory as `.comet.json` or to `${XDG_CONFIG_HOME}/cometary/config.json`. Omitting this means that the same defaults are used as in the original.

- To adjust the character limit of the scope, add the key `scopeInputCharLimit` with the desired limit
  - Default: 16
- To adjust the character limit of the message, add the key `commitInputCharLimit` with the desired limit
  - Default: 100
- To adjust the total limit of characters in the *resulting* commit message, add the key `totalInputCharLimit` with the desired limit
  - Adding this key overrides scope- and message-specific limits
- To allow typing beyond the character limit while still showing the count, add the key `overflowCharLimit` with the value `true`
  - Default: `false`
  - When enabled, the character count will turn orange when the limit is exceeded
- To adjust the order of the scope completion values (i.e. longer or shorter strings first), add the key `scopeOrderCompletion` with either `"ascending"` or `"descending"`
  - Default: `"descending"`
- To enable the storing of runtime statistics, add the key `storeRuntime` with the value `true`
  - Default: `false`
  - This will create a `stats.json` file next to the configuration file with aggregated statistics across days, weeks, months, and years
- To show the session runtime statistics after each commit, add the key `showRuntime` with the value `true`
  - Default: `false`
  - This will show `> Session: N seconds` after the commit was successful
- To show the all-time runtime statistics after each commit, add the key `showStats` with the value `true`
  - Default: `false`
  - To just show the all-time runtime statistics and quit the program, run the program with the `-s` flag
- To adjust the format of the statistics from seconds to hours or minutes, add the key `showStatsFormat` with either `"minutes"` or `"hours"`
  - Default: `"seconds"`
- To always show session runtime statistics as seconds but keep everything else as defined by `showStatsFormat`, add the key `sessionStatAsSeconds` with the value `true`
  - Default: `false`
- To use a custom color scheme, add the key `colorScheme` with the filename of your color scheme file
  - The color scheme file should be a JSON file placed adjacent to your configuration file (or in the current directory)
  - All fields are optional - any omitted values will fall back to defaults
  - Example color scheme file:
    ```json
    {
      "selectedItemColors": {
        "light": "#ff6b6b",
        "dark": "#98c379"
      },
      "versionStyle": {
        "light": "#9b9b9b",
        "dark": "#5c5c5c"
      },
      "selectedItemIndicator": "→"
    }
    ```
  - Available fields:
    - `titleTextStyle`: marginLeft
    - `titleStyle`: marginLeft
    - `itemStyle`: paddingLeft
    - `characterCountColors`: light, dark
    - `overflowCharColor`: light, dark
    - `selectedItemColors`: light, dark
    - `selectedItemStyle`: paddingLeft
    - `selectedItemPadded`: paddingLeft
    - `itemDescriptionStyle`: paddingLeft, faint
    - `paginationStyle`: paddingLeft
    - `helpStyle`: paddingLeft, paddingBottom
    - `quitTextStyle`: margin, marginTop, marginBottom, marginLeft
    - `versionStyle`: light, dark
    - `selectedItemIndicator`: string (e.g. ">", "→", "*")

There is also a `-m` flag that takes a string that will be used as the basis for a search among all commit messages. For example: if you're committing something of a chore and always just use the message "update dependencies", you can do `cometary -m update` (use quotation marks if argument to `-m` includes spaces) and Cometary will populate the list of possible messages with those that include "update", which can then be cycled through with the Tab key. This is similar to the search you could make with `git log --grep="update"`.

By default the `-m` flag behavior is set to only populate with possible messages that adhere to conventional commits, but this behavior can be changed by setting the `findAllCommitMessages` value in the configuration file as `true`.

## Acknowledgments

Couldn't have been possible without the work of [Liam Galvin](https://github.com/liamg).

## License

[MIT](/LICENSE)
