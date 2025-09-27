const fs = require("fs");
const path = require("path");

// Read the test.json file
const testJsonPath = path.join(__dirname, "js/src/app/test.json");
const testData = JSON.parse(fs.readFileSync(testJsonPath, "utf8"));

// Function to add code field to files
function addCodeToFiles(data) {
  if (Array.isArray(data)) {
    return data.map((item) => addCodeToFiles(item));
  } else if (data && typeof data === "object") {
    if (data.type === "FILE" && !data.code) {
      // Add appropriate code content based on file extension
      if (data.extension === "CSV") {
        data.code = `# Sample CSV data for ${data.name}\nDate,Open,High,Low,Close,Volume\n2024-01-01,100.00,105.00,98.00,102.00,1000000\n2024-01-02,102.00,108.00,101.00,106.00,1200000\n2024-01-03,106.00,110.00,104.00,108.00,1100000`;
      } else if (data.extension === "PY") {
        data.code = `# Python file: ${data.name}\nprint("Hello from ${data.name}")\n\n# Add your Python code here\n`;
      } else if (data.extension === "CSS") {
        data.code = `/* CSS file: ${data.name} */\nbody {\n  font-family: Arial, sans-serif;\n  margin: 0;\n  padding: 0;\n}\n\n/* Add your styles here */`;
      } else if (data.extension === "JS") {
        data.code = `// JavaScript file: ${data.name}\nconsole.log('Hello from ${data.name}');\n\n// Add your JavaScript code here\n`;
      } else if (data.extension === "TS") {
        data.code = `// TypeScript file: ${data.name}\nconsole.log('Hello from ${data.name}');\n\n// Add your TypeScript code here\n`;
      } else if (data.extension === "HTML") {
        data.code = `<!DOCTYPE html>\n<html>\n<head>\n  <title>${data.name}</title>\n</head>\n<body>\n  <h1>Hello from ${data.name}</h1>\n</body>\n</html>`;
      } else {
        data.code = `# File: ${data.name}\n# This is a ${data.extension} file\n# Add your content here\n`;
      }
    }

    // Recursively process subdirectories
    if (data.subDirectories) {
      data.subDirectories = data.subDirectories.map((item) =>
        addCodeToFiles(item)
      );
    }
  }

  return data;
}

// Process the data
const updatedData = addCodeToFiles(testData);

// Write the updated data back to the file
fs.writeFileSync(testJsonPath, JSON.stringify(updatedData, null, 2));

console.log("Successfully added code fields to all files in test.json");
