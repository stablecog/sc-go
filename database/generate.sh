#!/bin/bash
go run  entgo.io/ent/cmd/ent generate --feature sql/execquery --feature sql/modifier --feature sql/upsert ./ent/schema
