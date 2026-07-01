---
type: concept
title: {broken
description: The frontmatter YAML below the title fails to parse.
timestamp: 2026-01-01
tags: [example]
aliases: [a]
resource: https://example.com/malformed
---
# Malformed YAML

Only the `title` value is mutated into an unterminated YAML flow mapping, so the
frontmatter fails to parse and the page trips `okf-yaml-parse` alone.
