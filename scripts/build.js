const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

function build() {
    // Resolve paths relative to the parent directory of this script (workspace root)
    const rootDir = path.join(__dirname, '..');
    
    // 1. Path to package.json
    const packageJsonPath = path.join(rootDir, 'package.json');
    const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));

    // 2. Parse current version and increment the patch number
    const currentVersion = packageJson.version || '1.0.0';
    const versionParts = currentVersion.split('.');
    
    if (versionParts.length === 3) {
        const patch = parseInt(versionParts[2], 10);
        versionParts[2] = (isNaN(patch) ? 0 : patch + 1).toString();
    } else {
        // If version is not in major.minor.patch format, convert it to one
        while (versionParts.length < 3) {
            versionParts.push('0');
        }
        versionParts[2] = '1';
    }
    const newVersion = versionParts.join('.');

    // 3. Update package.json
    packageJson.version = newVersion;
    fs.writeFileSync(packageJsonPath, JSON.stringify(packageJson, null, 4) + '\n', 'utf8');

    // 4. Update src/version.go
    const versionGoPath = path.join(rootDir, 'src', 'version.go');
    const versionGoContent = `package main

// Version is the current version of the application.
// This file is auto-generated during the build process.
const Version = "${newVersion}"
`;
    fs.writeFileSync(versionGoPath, versionGoContent, 'utf8');

    console.log(`Incremented version from ${currentVersion} to ${newVersion}`);

    // 5. Run go build
    try {
        console.log('Building Go binary...');
        
        // Ensure bin directory exists
        const binDir = path.join(rootDir, 'bin');
        if (!fs.existsSync(binDir)) {
            fs.mkdirSync(binDir);
        }
        
        // Run go build in the workspace root directory
        execSync('go build -o bin/copy-pm-files.exe ./src', { 
            cwd: rootDir,
            stdio: 'inherit' 
        });
        console.log('Build successful!');
    } catch (error) {
        console.error('Build failed:', error);
        process.exit(1);
    }
}

build();
