---
type: source
title: A Source With an Invalid source_type
description: source_type value is outside the enum set.
timestamp: 2026-01-01
tags: [nlp]
aliases: [bad-source-type]
resource: https://example.com/x
authors: [Solo]
source_type: journal
doi: 10.1/x
---
# A Source With an Invalid source_type

`journal` is not a member of the source_type enum, so profile-field-enum fires.
