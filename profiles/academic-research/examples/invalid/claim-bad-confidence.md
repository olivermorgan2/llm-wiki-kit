---
type: claim
title: A Claim With an Invalid confidence
description: confidence value is outside the enum set.
timestamp: 2026-01-01
tags: [nlp]
aliases: [bad-confidence]
resource: https://example.com/x
confidence: certain
assessment: open
---
# A Claim With an Invalid confidence

## Evidence

An open claim carries no citation obligation.

## Counterevidence

None.

## Assessment

`certain` is not a member of the confidence enum, so profile-field-enum fires.
