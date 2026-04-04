import * as path from "node:path";
import * as vscode from "vscode";
import type { RepoDoctorReport } from "./analyzer";

export class RepoDoctorDiagnostics {
  private readonly collection: vscode.DiagnosticCollection;

  constructor() {
    this.collection = vscode.languages.createDiagnosticCollection("repodoctor");
  }

  public dispose(): void {
    this.collection.dispose();
  }

  public publish(workspacePath: string, report: RepoDoctorReport): void {
    const grouped = new Map<string, vscode.Diagnostic[]>();

    this.pushCircularDiagnostics(grouped, workspacePath, report.circularViolations ?? []);
    this.pushLayerDiagnostics(grouped, workspacePath, report.layerViolations ?? []);
    this.pushSizeDiagnostics(grouped, workspacePath, report.sizeViolations ?? []);
    this.pushGodObjectDiagnostics(grouped, workspacePath, report.godObjectViolations ?? []);

    this.collection.clear();
    for (const [filePath, diagnostics] of grouped.entries()) {
      this.collection.set(vscode.Uri.file(filePath), diagnostics);
    }
  }

  private pushCircularDiagnostics(grouped: Map<string, vscode.Diagnostic[]>, workspacePath: string, violations: Array<{ path?: string[]; hint?: string }>): void {
    for (const violation of violations) {
      if (!violation.path || violation.path.length === 0) {
        continue;
      }
      const cycle = violation.path.join(" -> ");
      this.add(grouped, this.resolvePath(workspacePath, violation.path[0]), this.withHint(`Circular dependency: ${cycle}`, violation.hint), vscode.DiagnosticSeverity.Error);
    }
  }

  private pushLayerDiagnostics(grouped: Map<string, vscode.Diagnostic[]>, workspacePath: string, violations: Array<{ from?: string; message?: string; hint?: string }>): void {
    for (const violation of violations) {
      if (!violation.from) {
        continue;
      }
      const message = this.withHint(violation.message ?? "Layer violation", violation.hint);
      this.add(grouped, this.resolvePath(workspacePath, violation.from), message, vscode.DiagnosticSeverity.Error);
    }
  }

  private pushSizeDiagnostics(grouped: Map<string, vscode.Diagnostic[]>, workspacePath: string, violations: Array<{ file?: string; function?: string; lines?: number; threshold?: number; hint?: string }>): void {
    for (const violation of violations) {
      if (!violation.file) {
        continue;
      }
      const base = violation.function
        ? `Function '${violation.function}' has ${violation.lines ?? 0} lines (threshold: ${violation.threshold ?? 0})`
        : `File has ${violation.lines ?? 0} lines (threshold: ${violation.threshold ?? 0})`;
      this.add(grouped, this.resolvePath(workspacePath, violation.file), this.withHint(base, violation.hint), vscode.DiagnosticSeverity.Warning);
    }
  }

  private pushGodObjectDiagnostics(grouped: Map<string, vscode.Diagnostic[]>, workspacePath: string, violations: Array<{ struct?: string; file?: string; fields?: number; methods?: number; hint?: string }>): void {
    for (const violation of violations) {
      if (!violation.file) {
        continue;
      }
      const base = `Struct '${violation.struct ?? "unknown"}' has ${violation.fields ?? 0} fields and ${violation.methods ?? 0} methods`;
      this.add(grouped, this.resolvePath(workspacePath, violation.file), this.withHint(base, violation.hint), vscode.DiagnosticSeverity.Warning);
    }
  }

  private add(grouped: Map<string, vscode.Diagnostic[]>, filePath: string, message: string, severity: vscode.DiagnosticSeverity): void {
    const diagnostics = grouped.get(filePath) ?? [];
    const diagnostic = new vscode.Diagnostic(new vscode.Range(0, 0, 0, 1), message, severity);
    diagnostic.source = "RepoDoctor";
    diagnostics.push(diagnostic);
    grouped.set(filePath, diagnostics);
  }

  private withHint(message: string, hint?: string): string {
    if (!hint || hint.trim() === "") {
      return message;
    }
    return `${message}. Hint: ${hint}`;
  }

  private resolvePath(workspacePath: string, candidate: string): string {
    if (path.isAbsolute(candidate)) {
      return candidate;
    }
    return path.join(workspacePath, candidate);
  }
}
