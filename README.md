
# Central Cyclone 🌪️

Central Cyclone is a centralized SBOM (Software Bill of Materials) generation service built around [cdxgen](https://github.com/CycloneDX/cdxgen). It can be configured to analyze and upload the results for multiple repos and multiple targets, so you do not have to create a pipeline for each repository. Similar to how Renovate can be configured to manage multiple repositories.

[<img  src="assets/Central-Cyclone.drawio.svg">]()


## Table of Contents
- [Features](#features)
- [Usage](#usage)
    - [Configuration](#configuration)
    - [Commands](#commands)
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
                    "projectId": "2fbbfb99-132e-4e8d-b253-4aa8d58aa505",
                    "type": "node",
                    "directory": "web"
                }, {
                    "projectId": "2fbbfb99-132e-3d8d-b253-4aa8d58aa505",
                    "type": "java"
                }
            ]
            
        }
    ],
    ""applications": [
        {
            "name": "My-App",
            "type": "node",
            "projects": [
                {
                    "name": "My-App - Dev",
                    "version": "Dev",
                    "isLatest": true
                },
                {
                    "name": "My-App - Prod",
                    "version": "Prod",
                    "isLatest": false
                }
            ]
        }
    ]
}

```
The `dependencyTrack` section in your configuration file is **mandatory**, as is setting the `DEPENDENCYTRACK_API_KEY` environment variable. For more details, see the  [Environment Variables](#environment-variables) section.

You can configure multiple *targets* for a single repository. This can be useful for a monorepo, where different programming languages or projects are managed under a single repository. You can find all supported targets in the [Cdxgen documentation](https://cyclonedx.github.io/cdxgen/#/PROJECT_TYPES).
If you project contains multiple subprojects of the same type. You can specify the subdir within the repo using the optional `directory` property.


The new block `applications` is optional and can be used to define application. An application can contain multiple *Projects*. Each project represents a project in DependencyTrack.
This concept will be used in future updates to enable an GitOps mode in which central cyclone will monitor you gitops repo(s) and create sboms for the deployed versions on you environments.

### Commands


#### Global Parameter

|Parameter| Shortcut| Description|
|-|-|-|
|`--config`| `-c`| Path to the config file|



#### Analyze configured repos
The `analyze` command clones and analyzes all configured repositories for their defined targets. The resulting SBOMs are either saved under `~/.central-cyclone/workspace/sboms` or directly uploaded.


```
analyze 
```
- `-c path-to-config`: Path to your configuration JSON file.
- `--upload`: Optional, uploads the resulting sboms instead of saving them.

#### Upload
The upload command can be used to upload the sbom files resulting from the analyze command. This can be useful in restricted network environments. You can use a two stage pipeline to first analyze the projects on a cloud agent and use a self hosted agent to upload the reuslting sboms.

It is important, that Central Cyclone can only upload sboms created by Central Cyclone. Otherwise, Central Cyclone is not able to match the sbom file with the corresponding project.

```
upload
```

- `-c path-to-config`: Path to your configuration JSON file.
- `--sboms-dir`: Required, path to dir containing all the sboms to upload.
`
#### Sync Projects with DependencyTrack
You can use central cyclone to create and sync DependencyTrack projects. However, this feature is not a configuration as code solution. As described in the config, it will only sync projects defined for applications. This will later be used for the GitOps mode of central cycline.

To trigger the sync use this command:

```
dt projects sync
```
- `-c path-to-config`: Path to your configuration JSON file

The command will only create projects, if they don't exist. Existing projects are not deleted and not updated.

### Docker Image
We provide an official docker image under the packages section of GitHub. It's recommended to use the docker image to run Central-Cyclone as it already includes all dependencies such as `git` and `cdxgen`.

To use it, create a config as stated above  and mount it into the container. Just make sure, that the path given to the analyze command matches the mounted one:
```
docker run \
-v ./myConfig.json:/config/config.json \
-e DEPENDENCYTRACK_API_KEY=MY_API_KEY \
ghcr.io/bjarnerentz/central-cyclone:latest analyze -c /config/config.json
```

The easiest way to extract the sboms of the `analyze` command is to mount the workfolder of central cyclone.
The workfolder is located under the home directory, for the current dockre image this is `\root`.


```
docker run \
-v ./myConfig.json:/config/config.json \
-v ./sboms:/root/.central-cyclone/workfolder/sboms \
-e DEPENDENCYTRACK_API_KEY=MY_API_KEY \
ghcr.io/bjarnerentz/central-cyclone:latest analyze -c /config/config.json
```


### Cloning Private Repositories
**Supported Platforms:**
Currently, this feature is tested with GitHub and Azure DevOps. If you successfully use it with another platform, please let us know!

**URL Transformation Examples:**

GitHub:
https://github.com/<User>/<Repo>.git
⟶ https://<Token>@github.com/<User>/<Repo>.git
Azure DevOps:
https://dev.azure.com/<Org>/<Project>/_git/<Repo>
⟶ https://<Token>@dev.azure.com/<Org>/<Project>/_git/<Repo>

**Note:**

For GitHub, use fine-grained personal access token. Select the repos you want to clone and add the *Contents* permission as Read-only.
For Azure DevOps, use a PAT with "Code" read permissions.
If you encounter any issues or need support for other platforms, please open an Issue so we can improve future releases.

Also check the documentation for available pipeline variables such as [`System.AccessToken`](https://learn.microsoft.com/de-de/azure/devops/pipelines/build/variables?view=azure-devops&tabs=yaml#systemaccesstoken) in Azure Devops to prevent using long running tokens.

## Environment Variables
- `DEPENDENCYTRACK_API_KEY` (required): API key for authenticating with Dependency-Track.
- `GIT_TOKEN` (optional) can be set to clone private repositories.

The API key only needs the BOM-Upload permissions for the projects. Central Cyclone will not create projects for you within DependencyTrack.

## Example
See `exampleConfig.json` for a minimal working configuration.

## Roadmap
- Support GitOps: Automaticly create SBOMs for the deployed version of you apps.


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
This project was created with the support of AI. Feel free to let AI assist you with pull requests, but please review the changes yourself.