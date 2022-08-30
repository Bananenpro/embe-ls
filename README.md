# embe-ls

An [LSP](https://microsoft.github.io/language-server-protocol) implementation for [embe](https://github.com/Bananenpro/embe).

## Features

- [x] diagnostics
- [x] code completion
- [x] snippets
- [ ] documentation on hover
- [ ] symbol rename

## Installation

[Build](#building) the `embe-ls` binary and place it somewhere in your PATH.

### VS Code

Install the [vscode-embe](https://github.com/Bananenpro/vscode-embe#installation) extension.

### Neovim

Install the [vim-embe](https://github.com/Bananenpro/vim-embe#installation) plugin for syntax highlighting and indentation.

#### coc

In [`coc-settings.json`](https://github.com/neoclide/coc.nvim/wiki/Language-servers#register-custom-language-servers):
```json
{
  "languageserver": {
    "embe-ls": {
      "command": "embe-ls",
      "filetypes": ["embe"],
      "rootPatterns": [".git/"]
    }
  }
}
```

#### lspconfig

In `init.lua`:
```lua
local lspconfig = require('lspconfig')
local configs = require('lspconfig.configs')
configs.embe = {
  default_config = {
    cmd = { "embe-ls" },
    root_dir = lspconfig.util.root_pattern('.git'),
    filetypes = { 'embe' },
    init_options = {
      command = { 'embe-ls' },
    },
  },
}
lspconfig.embe.setup{}
```

## Building

### Prerequisites

- [Go](https://go.dev) 1.19+

```
git clone https://github.com/Bananenpro/embe-ls
cd embe-ls
go build
```

## License

Copyright (c) 2022 Julian Hofmann

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
