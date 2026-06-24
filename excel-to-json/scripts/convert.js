#!/usr/bin/env node
/**
 * Convert an Excel worksheet to a JSON array.
 *
 * Usage:
 *   node convert.js <excel_file> [--sheet <name>] [--output <file>]
 *
 * The xlsx package is auto-installed if missing.
 */

const { execSync } = require("child_process");
const path = require("path");
const fs = require("fs");

function ensureXlsx() {
  try {
    require.resolve("xlsx");
  } catch {
    console.error("xlsx package not found — installing...");
    execSync("npm install xlsx", {
      cwd: path.dirname(__filename),
      stdio: "inherit",
    });
  }
  return require("xlsx");
}

function parseArgs(argv) {
  const args = { excelFile: null, sheet: null, output: null };
  const rest = argv.slice(2);
  for (let i = 0; i < rest.length; i++) {
    if (rest[i] === "--sheet" || rest[i] === "-s") {
      args.sheet = rest[++i];
    } else if (rest[i] === "--output" || rest[i] === "-o") {
      args.output = rest[++i];
    } else if (!args.excelFile) {
      args.excelFile = rest[i];
    }
  }
  return args;
}

function main() {
  const args = parseArgs(process.argv);

  if (!args.excelFile) {
    console.error(
      "Usage: node convert.js <excel_file> [--sheet <name>] [--output <file>]"
    );
    process.exit(1);
  }

  if (!fs.existsSync(args.excelFile)) {
    console.error(`Error: file not found: ${args.excelFile}`);
    process.exit(1);
  }

  const XLSX = ensureXlsx();

  let workbook;
  try {
    workbook = XLSX.readFile(args.excelFile, { cellDates: true });
  } catch (err) {
    console.error(`Error reading Excel file: ${err.message}`);
    process.exit(1);
  }

  const sheetName = args.sheet ?? workbook.SheetNames[0];
  if (!workbook.SheetNames.includes(sheetName)) {
    console.error(
      `Error: sheet "${sheetName}" not found.\nAvailable sheets: ${workbook.SheetNames.join(", ")}`
    );
    process.exit(1);
  }

  const sheet = workbook.Sheets[sheetName];
  const records = XLSX.utils.sheet_to_json(sheet, {
    defval: null,   // empty cells → null
    raw: true,      // preserve native number/boolean types; cellDates:true handles dates
  });

  const json = JSON.stringify(records, null, 2);

  if (args.output) {
    fs.writeFileSync(args.output, json, "utf-8");
    console.log(`JSON written to ${args.output} (${records.length} records)`);
  } else {
    process.stdout.write(json + "\n");
  }
}

main();
