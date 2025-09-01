
# Central Cyclone 🌪️

Central Cyclone is a centralized SBOM (Software Bill of Materials) generation service built around [cdxgen](https://github.com/CycloneDX/cdxgen). It can be configured to analyze and upload the results for multiple repos and multiple targets, so you do not have to create a pipeline for each repository. Similar to how Renovate can be configured to manage multiple repositories.


## Table of Contents
- [Features](#features)
- [Usage](#usage)
    - [Configuration](#configuration)
    - [Command](#command)
- [Environment Variables](#environment-variables)
- [Example](#example)
- [Roadmap](#roadmap)
- [Contributing & Support](#contributing--support)
- [Development Setup](#development-setup)
    - [DevContainer](#devcontainer)
    - [Local Machine](#local-machine)
- [AI Disclaimer](#ai-disclaimer)

## Features
- Centralized SBOM generation for multiple repositories/projects
- Upload to [DependencyTrack](https://dependencytrack.org)
- Configuration-driven: manage all targets in a single config file
- Command-line interface for easy automation

## Usage

### Configuration
Define your targets and settings in a JSON config file. See `exampleConfig.json` for a sample configuration. It looks like this 
```json
{
    "dependencyTrack": {
        "url": "http://apiserver:8080"
    },
    "repositories": [
        {
            "url": "https://github.com/BjarneRentz/obsidian-gemini-generator.git",
            "targets": [
                {
                    "projectId": "obsidian-gemini-generator-node",
                    "type": "node"
                }, {
                    "projectId": "obsidian-gemini-generator-java",
                    "type": "java"
                }
            ]
            
        }
    ]
}

```
The `dependencyTrack` section in your configuration file is **mandatory**, as is setting the `DEPENDENCYTRACK_API_KEY` environment variable. For more details, see the  [Environment Variables](#environment-variables) section.

You can configure multiple targets for a single repository. This can be useful for a monorepo, where different programming languages or projects are managed under a single repository.


### Command
```
analyze -c path-to-config
```
- `-c path-to-config`: Path to your configuration JSON file.

## Environment Variables
- `DEPENDENCYTRACK_API_KEY` (required): API key for authenticating with Dependency-Track.

The API key only needs the BOM-Upload permissions for the projects. Central Cyclone will not create projects for you within DependencyTrack.

## Example
See `exampleConfig.json` for a minimal working configuration.

## Roadmap
- Official Docker image for easy use
- Git access token support to use Central Cyclone on private repos.


## Contributing & Support
For questions, issues, or contributions, please open an issue or pull request on GitHub.


### Development Setup

#### DevContainer
This project comes with a DevContainer setup that ships all required dependencies:

- git
- cdxgen
- DependencyTrack

Upon the first start, you can log in to DependencyTrack at `http://localhost:8080` with username `admin` and password `admin`. You are prompted to change the default password for the `admin` user afterwards. The DevContainer is configured to use a volume for DependencyTrack and thus will persist the new password.

Next, create a project and a new team with an API key to be used by Central Cyclone. Further details on this can be found in the official [DependencyTrack documentation](https://docs.dependencytrack.org).

#### Local Machine
If you do not want to use the DevContainer, make sure that Central Cyclone has access to the following tools via your `PATH`:
- git
- cdxgen

and can reach a DependencyTrack instance.


## AI Disclaimer
This project was created with the support of GitHub Copilot. Feel free to let AI assist you with pull requests, but please review the changes yourself.