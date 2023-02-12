# stablecog/sc-go/database

This hosts database-related items including Postgres, Redis,  Supabase, and Meili.

## Appendix

- `./ent` - primary auto-generated files created by the [ent ORM](https://entgo.io/)
- `./generate.sh` - Needed to re-generate the files in `./ent` when changes are made to the schema `./ent/schema`
- `./repository` - Collection of functions used to access and modify the SQL database.
## Design Philosphy

The SQL database should not be accessed using the ent or supabase client directly. Our design philosphy is that database access should occur only within this package, and other packages should reference this package functions when interacting with the database.

This helps our codebase stay organized, testable, and reliable.

SQL interactions should happen within `./repository` (excluding supabase), each file in repository should be associated with a schema in `./ent/schema`. Excluding some special considerations such as stored procedures, direct access, etc.