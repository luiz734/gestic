# gestic

A diff tool for restic snapshots.

## Problem
- Your last [restic](https://github.com/restic/restic) snapshot is `snapshot_old`
- You make a new snapshot `snapshot_new`
- `snapshot_new` have +800MB on size compared to `snapshot_old`
- You want to find what changed

Sometimes, we install some program or download a file that is not worth to keep on the snapshots, like VM files.
This tool helps you quickly find what file or directory is causing the change in size.


## Usage
**Make sure your restic repository is mounted before run it.**

`restic mount /mnt/YOUR_MOUNT_POINT`

After mount the repo, you can use the tool:

`gestic --repo /mnt/YOUR_RESTIC_REPO --mount /mnt/YOUR_MOUNT_POINT`

Use the help on screen to move around and compare snapshots.


## Installation
- Clone the repo
- Use the install script or compile it manually

