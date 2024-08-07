const vscode = require('vscode');

function activate(context) {
    let disposable = vscode.languages.registerDocumentLinkProvider({ scheme: 'file' }, {
        provideDocumentLinks(document) {
            const links = [];
            const regex = /mo\/[a-zA-Z0-9_-]+/g;
            
            for (let i = 0; i < document.lineCount; i++) {
                const line = document.lineAt(i);
                let match;
                
                while ((match = regex.exec(line.text)) !== null) {
                    const start = new vscode.Position(i, match.index);
                    const end = new vscode.Position(i, match.index + match[0].length);
                    const range = new vscode.Range(start, end);
                    const uri = vscode.Uri.parse(`http://${match[0]}`);
                    
                    links.push(new vscode.DocumentLink(range, uri));
                }
            }
            
            return links;
        }
    });

    context.subscriptions.push(disposable);
}

function deactivate() {}

module.exports = {
    activate,
    deactivate
};