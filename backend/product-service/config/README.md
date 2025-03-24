# Configuration Folder Documentation

## Overview
The `config` folder contains the configuration settings for the product service. These settings are essential for defining how the service operates, including server settings, database connections, and secret management.

## Structure
The main configuration file is `config.go`, which defines the structure of the configuration settings using Go structs. The configuration is divided into three main sections:

- **ServerConfig**: Contains settings related to the server, such as host and port.
- **DatabaseConfig**: Contains settings for connecting to the database, such as the database URL, username, and password.
- **SecretsConfig**: Contains sensitive information, such as API keys and secret tokens, which are loaded from environment variables for security reasons.

## Configuration Management
- **Loading Configuration**: The configuration settings are typically loaded using a configuration management library that supports `mapstructure` tags. This allows the configuration to be loaded from various sources, such as JSON, YAML, or environment variables.
- **Environment Variables**: Sensitive information, such as secrets, should be managed using environment variables. This ensures that sensitive data is not hardcoded in the source code or configuration files.

## Best Practices
- **Security**: Ensure that sensitive information is not exposed in version control systems. Use environment variables or secret management services to handle sensitive data.
- **Version Control**: Keep configuration files under version control, but exclude files containing sensitive information, such as `.env` files, by adding them to `.gitignore`.
- **Documentation**: Keep this documentation up-to-date to help developers understand how to manage and use configuration settings effectively.

## Usage
Developers should refer to this documentation when adding or modifying configuration settings. Ensure that any changes to the configuration structure are reflected in this document.