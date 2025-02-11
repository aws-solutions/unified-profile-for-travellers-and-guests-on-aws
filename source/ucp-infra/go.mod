// AWS CDK node modules are packaged with references to Go packages. These Go packages are detected "go mod tidy" and result in unnecessary packages being
// added to the base "go.mod" file.
//
// Adding an empty go.mod to any directory where those AWS CDK node modules may be present prevents that behavior
// https://go.dev/wiki/Modules#can-an-additional-gomod-exclude-unnecessary-content-do-modules-have-the-equivalent-of-a-gitignore-file
//
// Example of a Go package that would mistakenly be referenced
// https://github.com/aws/aws-cdk/tree/main/packages/aws-cdk/lib/init-templates/app/go