# Mo Links

Mo Links is a Visual Studio Code extension that automatically converts text patterns like `mo/some_text-examples` into clickable links. This extension works across all programming languages and file types in VSCode. Enables easy embedding of mo-links within your code comments for easy typing and easy clicking!

## Features

- Automatically detects and converts `mo/[text]` patterns into clickable links
- Works in all file types and programming languages
- Assumes that Chrome is your default browser, with the Mo Links browser extension

## Installation

1. Open Visual Studio Code
2. Press `Ctrl+P` (or `Cmd+P` on macOS) to open the Quick Open dialog
3. Type `ext install mo-links` and press Enter
4. Restart VS Code after the installation

Alternatively, you can install the extension directly from the Visual Studio Code Marketplace.

## Usage

After installation, the extension will automatically activate for all files. No additional configuration is required.

1. Open any file in VSCode
2. Type or paste text containing patterns like `mo/example-text`
3. The matching text will automatically become a clickable link

## Configuration

Currently, the extension does not have any configurable settings. The link destination is hardcoded to `http://mo/[text]`. To change this, you'll need to modify the source code.

## Contributing

Contributions to the Mo Link extension are welcome! Please feel free to submit a Pull Request.

## Support

If you encounter any problems or have any suggestions, please open an issue on the GitHub repository. [https://github.com/jacksonStone/molinks_extensions](https://github.com/jacksonStone/molinks_extensions)

## Acknowledgments

- Thanks to the VSCode team for providing excellent documentation and APIs for extension development.

## Disclaimer

This extension is provided as-is, without any warranty. Use at your own risk.