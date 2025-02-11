#!/bin/bash

echo "Starting dependency setup..."

cd ..
ROOT_DIR="$(pwd)"
SOURCE_DIR="$ROOT_DIR/source"

# Setup Go project
echo "Setting up Go project..."
if [ ! -f "go.mod" ]; then
    echo "Initializing Go module in source directory"
    go mod init tah/upt
fi
echo "Tidying Go modules"
go mod tidy

echo "Vendoring dependencies"
go mod vendor

# Setup Node.js projects
echo "Setting up Node.js projects..."
find "$SOURCE_DIR" "$ROOT_DIR/deployment/cdk-solution-helper" -name "package.json" -not -path "*/node_modules/*" | while read package_json; do
    project_dir=$(dirname "$package_json")
    echo "Installing dependencies in $project_dir"
    cd "$project_dir"
    rm -rf node_modules
    npm install
done

cd $SOURCE_DIR

# Setup Python dependencies
echo "Setting up Python dependencies..."
find . -name "requirements.txt" -not -path "*/node_modules/*" -type f | while read -r req_file; do
    project_dir=$(dirname "$req_file")
    echo "Installing Python dependencies in $project_dir"
    
    # Verify directory exists and is accessible
    if [ ! -d "$project_dir" ]; then
        echo "Directory $project_dir not found, skipping..."
        continue
    fi
    
    cd "$project_dir" || continue
    
    # Verify requirements.txt exists
    if [ ! -f "requirements.txt" ]; then
        echo "requirements.txt not found in $project_dir, skipping..."
        cd "$SOURCE_DIR" || exit
        continue
    fi
    
    # Create virtual environment if it doesn't exist
    if [ ! -d "venv" ]; then
        echo "Creating virtual environment in $project_dir"
        python3 -m venv venv
    fi
    
    # Activate virtual environment
    source venv/bin/activate || {
        echo "Failed to activate virtual environment in $project_dir"
        cd "$SOURCE_DIR" || exit
        continue
    }
    
    # Install requirements
    if pip3 install -r requirements.txt; then
        echo "Successfully installed requirements in $project_dir"
    else
        echo "Failed to install requirements in $project_dir"
    fi
    
    # Deactivate virtual environment
    deactivate
    
    cd "$SOURCE_DIR" || exit
done

cd $SOURCE_DIR
if [ ! -d "z-coverage" ]; then
    mkdir z-coverage
fi

echo "Setup complete!"
