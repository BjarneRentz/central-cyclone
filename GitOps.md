# Central Cyclone GitOps Mode

The GitOps mode enables the automated creation of sboms for different deployed environements depending on the state of you GitOps repository. This has the advantage, that you sboms are always up to date while you teams still do not have to adapt any pipeline or release process. Central Cyclone will use you single point of truth - the GitOps repo - to get the deployed version and create a matching SBOM for it.


## Config

ToDo: Rename Version to environment in DependencyTrack Projects to reduce confusion

The config is quite complex on the first sight. This is due to the lose coupling betweens the applications themself, their repositories, the GitOps repo(s) and the corresponding DependencyTrack Project for each app in each of its versions. For a better understanding, we first define the terminology.

**Application**: An application is a standalone software project that can be shipped independant. *Services* can be built around *Applications*. For example a *Basket-Service* can be made out of a *Basket-Service Backend* and *Basket-Service Frontend*. The Frontend and Backend can be deployed differently, be in different repositories or a shared one.

**DependencyTrack Project**: A DependencyTrack Project holds the SBOM and metrics for a single *Application* deployed in a specific environment called *version* in DependencyTrack like *Dev*, *Staging* or *Prod*. Hence, there are separate DependencyTrack Projects for *Basket-Service Backend (Dev)*, *Basket-Service Backend (Prod)*, *Basket-Service Frontend (Dev)* and so on. It's recommended to use any form of IaC like Terraform for creation of all the projects. The important part is, that each combination of *Application* and *Environment* or *version* in terms of DependencyTrack has its own *DependencyTrack Project*.

**ApplicationRepos**: As already mentioned, there are many different ways to structure *Applications* in git repositories such as mono- or multirepo. Hence, a single *ApplicationRepo* can contain one or multiple *Applications*, identified by their name.


**GitOps Repos**: These represent you exsiting GitOps repos used by ArgoCD, Flux or any other tool of your choice. Our goal is to extract the deployed version of each *Application* that is defined by it. To get the version, a *VersionIdentifier* is configured that tells Central Cyclone under which filepath and yamlpath it can get the deployed version of *Basket-Service Backend*.

```json

"applications": [
    {
        "name": "Basket-Service Backend",
        "dependencyTrackProjects": [
            {
                "name": Basket-Service Backend (Dev),
                "version": "Dev",
                "projectId": "ksdf-1231-sdf-13-"
            },
            {
                "name": Basket-Service Backend (Staging),
                "version": "Staging",
                "projectId": "sjkdfs-234-sd-sdfs-"
            }
        ],
        "name": "Order-Service Backend",
        "dependencyTrackProjects": [
            {
                "name": Order-Service Backend (Dev),
                "version": "Dev", // Required to match the gitops app with its environement to the corresponding DependencyTrack project
                "projectId": "ksdf-1231-sdf-13-"
            },
            {
                "name": Order-Service Backend (Staging),
                "version": "Staging",
                "projectId": "sjkdfs-234-sd-sdfs-"
            }
        ]
    }
]


"gitOpsRepos": [
    {
        "url": "https://github.com/my-git-ops-repo1",
        "gitOpsApplications": [
            {
                "application": "Basket-Service Backend",
                "versionIdentifiers": [
                    {
                        "environment": "Dev",
                        "filePath": "apps/basket-service/dev/values.yaml",
                        "yamlPath": "backend.image.tag"
                    },
                    {
                        "environment": "Staging",
                        "filePath": "apps/basket-service/staging/values.yaml",
                        "yamlPath": "backend.image.tag"
                    }
                ]
            },
            {
                "application": "Order-Service Backend",
                "versionIdentifiers": [
                    {
                        "environment": "Dev",
                        "filePath": "apps/order-service/dev/values.yaml",
                        "yamlPath": "backend.image.tag"
                    },
                    {
                        "environment": "Staging",
                        "filePath": "apps/order-service/staging/values.yaml",
                        "yamlPath": "backend.image.tag"
                    }
                ]
            }
        ]
    }
],
"applicationRepos": [
    {
        "applications": [ 
            "Basket-Service Backend", "Basket-Service Frontend"
            ],
        "repoUrl": "https://github.com/basket-service"
    },
    {
        "applications": ["Order-Service Backend"],
        "repoUrl": "https://github.com/order-service
    }
]

```


### Current Pain Points
- Everything is held together by the application(name) like "Basket-Service Backend". On the one hand, this is useful as it allows a lose configuration and extensions like including "applicationImages" for example in order to also scan images of the corresponding apps-
- Linking with DependencyTrack Projects, this should be as flat as possible as different users may have different setups like collection projects or not.
- Intransparent Linking. To get the corresponding DependencyProject for a given application and the deployed environment. You first search the corresponding app via the name and then select the DependencyTrack project of it via the environment. Thus, application and environment both have to match.