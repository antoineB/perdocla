Cli to add your personnal document to an sqlite database.

Like https://docs.paperless-ngx.com/


```
perdocla [-db mydocs.db] subcommand

subcommand:

- list [-tags invoices,isp] [-date 2025-01-13] [-mime application/pdf] [-filename invoices.pdf] [-search text]
List the document and filter depending on options

- add [-tags invoices,isp] /home/antoineb/documents/2025_invoice.pdf
Add the new document to the database
If the -tags options spécify tags associated to the document

- get [-tags invoices,isp] [-date 2025-01-13] [-output 2025_invoice.pdf] id
Get the document infos from id
If the -tags option is set the document is set to the specific tags.
The same for -date.
The options -output will write the content of the document to the given file, it is the only way of viewing the actual document.

- createdb
Create the initial database use -db option to spécifier where
```
