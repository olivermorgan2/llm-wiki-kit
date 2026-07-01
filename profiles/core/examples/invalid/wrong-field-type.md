---
type: concept
title: [this, should, be, a, scalar, string]
description: A page whose title has the wrong YAML type (a sequence, not a string).
timestamp: 2026-01-01
tags: [example]
aliases: [a]
resource: https://example.com/wrong-field-type
---
# Wrong Field Type

Only `title`'s YAML type is mutated (scalar string to sequence), so the page
trips `core-field-type` alone; the required-title rule stays silent because the
field is present.
