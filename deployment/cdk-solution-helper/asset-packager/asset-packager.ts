// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
import archiver from "archiver";
import { createWriteStream } from "fs";
import { lstat, readdir, rename } from "fs/promises";
import path from "path";

/**
 * @description Class to help with packaging and staging cdk assets
 * on solution internal pipelines
 */
export class CDKAssetPackager {
  constructor(private readonly assetFolderPath: string) {}

  /**
   * @description get cdk asset paths
   * All cdk generated assets are prepended with "asset"
   */
  async getAssetPaths() {
    try {
      const allFiles = await readdir(this.assetFolderPath);
      const assetFilePaths = allFiles
        .filter((file) => file.includes("asset"))
        .map((file) => path.join(this.assetFolderPath, file));
      return assetFilePaths;
    } catch (err) {
      console.error(err);
      return [];
    }
  }

  /**
   * @description creates zip from folder
   * @param folderPath
   */
  async createAssetZip(folderPath: string): Promise<void> {
    const isDir = (await lstat(folderPath)).isDirectory();
    if (isDir) {
      const zipName = `${folderPath.split("/").pop()}.zip`;
      const outputPath = path.join(this.assetFolderPath, zipName);

      const output = createWriteStream(outputPath);
      const archive = archiver.create("zip", {
        zlib: { level: 6 },
      });

      archive.on("warning", (err) => {
        if (err.code === "ENOENT") {
          console.warn("Archive warning:", err);
        } else {
          throw err;
        }
      });

      archive.on("error", (err) => {
        throw err;
      });

      output.on("error", (err) => {
        throw err;
      });

      // Pipe archive data to the output file
      archive.pipe(output);

      // Add the entire directory to the archive
      archive.directory(path.join(folderPath, "./"), false);

      // Finalize the archive and wait for completion
      archive.finalize();

      // Wait for the archive to finish
      await new Promise<void>((resolve) => {
        output.on("close", resolve);
      });
    }
  }

  /**
   * @description moves zips to staging output directory in internal pipelines
   * @param outputPath
   */
  async moveZips(outputPath: string) {
    const allFiles = await readdir(this.assetFolderPath);
    const allZipPaths = allFiles.filter(
      (file) => path.extname(file) === ".zip"
    );
    for (const zipPath of allZipPaths) {
      const hashPart = zipPath.split("asset.").pop();
      if (!hashPart) {
        throw new Error(`Unexpected path: {${zipPath}}`);
      }
      await rename(
        path.join(this.assetFolderPath, zipPath),
        path.join(outputPath, hashPart)
      );
      // remove cdk prepended string "asset.*"
    }
  }
}
