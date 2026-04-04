import { execFile } from "node:child_process";

export interface RepoDoctorReport {
  circularViolations?: Array<{ path?: string[]; severity?: string; hint?: string }>;
  layerViolations?: Array<{ from?: string; to?: string; message?: string; hint?: string }>;
  sizeViolations?: Array<{ file?: string; function?: string; lines?: number; threshold?: number; hint?: string }>;
  godObjectViolations?: Array<{ struct?: string; file?: string; fields?: number; methods?: number; hint?: string }>;
}

export class RepoDoctorAnalyzer {
  public async run(binaryPath: string, workspacePath: string): Promise<RepoDoctorReport> {
    const output = await this.exec(binaryPath, ["analyze", "-path", workspacePath, "-format", "json"]);
    return this.parseReport(output);
  }

  private exec(binaryPath: string, args: string[]): Promise<string> {
    return new Promise((resolve, reject) => {
      execFile(binaryPath, args, { maxBuffer: 10 * 1024 * 1024 }, (error, stdout, stderr) => {
        if (error) {
          const details = stderr && stderr.trim() !== "" ? stderr.trim() : error.message;
          reject(new Error(details));
          return;
        }
        resolve(stdout);
      });
    });
  }

  private parseReport(raw: string): RepoDoctorReport {
    const trimmed = raw.trim();
    if (trimmed === "") {
      throw new Error("repodoctor returned empty JSON output");
    }
    return JSON.parse(trimmed) as RepoDoctorReport;
  }
}
