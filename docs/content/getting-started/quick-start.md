---
title: "Quick start"
description: "Run your first weread command."
weight: 30
---

Once `weread` is on your `PATH`:

```bash
weread --help       # see the command tree
weread version      # build info
```

This is a fresh scaffold, so the command tree is just `version` for now. Add
your first real command in `cli/`, build on the `weread` library package,
and document it here.

A good first command usually fetches one thing and prints it as JSON, so the
output pipes straight into `jq` and the rest of your tools.
