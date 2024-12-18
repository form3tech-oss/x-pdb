# Release Process

X-PDB release process is performed by the `Create Release` GHA Workflow.

The workflow needs the version of the release, from which branch it will be generated and if it will be a pre-release.

This process will perform the following actions:

1. Build and Publish X-PDB Docker images into GHCR.
2. Package and Publish the HelmChart into GHCR and Github Pages.
3. Create a Github Release with the changelog since the last release.
