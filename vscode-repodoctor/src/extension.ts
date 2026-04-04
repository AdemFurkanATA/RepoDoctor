import * as vscode from "vscode";
import { RepoDoctorAnalyzer } from "./analyzer";
import { RepoDoctorDiagnostics } from "./diagnostics";

export function activate(context: vscode.ExtensionContext): void {
  const analyzer = new RepoDoctorAnalyzer();
  const diagnostics = new RepoDoctorDiagnostics();

  const runAnalysis = async (): Promise<void> => {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
      vscode.window.showInformationMessage("RepoDoctor: open a workspace folder to run analysis.");
      return;
    }

    const config = vscode.workspace.getConfiguration("repodoctor");
    const binaryPath = config.get<string>("binaryPath", "repodoctor");

    try {
      const report = await analyzer.run(binaryPath, workspaceFolder.uri.fsPath);
      diagnostics.publish(workspaceFolder.uri.fsPath, report);
      vscode.window.setStatusBarMessage("RepoDoctor analysis completed", 2000);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Unknown error";
      if (message.includes("ENOENT")) {
        vscode.window.showErrorMessage("RepoDoctor binary not found. Install RepoDoctor and ensure it is available in PATH.");
        return;
      }
      vscode.window.showErrorMessage(`RepoDoctor analysis failed: ${message}`);
    }
  };

  context.subscriptions.push(diagnostics);
  context.subscriptions.push(vscode.commands.registerCommand("repodoctor.analyzeWorkspace", runAnalysis));

  context.subscriptions.push(
    vscode.workspace.onDidSaveTextDocument(async () => {
      const runOnSave = vscode.workspace.getConfiguration("repodoctor").get<boolean>("runOnSave", true);
      if (runOnSave) {
        await runAnalysis();
      }
    })
  );
}

export function deactivate(): void {}
