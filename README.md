DB tables:

TestStatus - id, status
Dog - id, name, gender, teststatus
Relationship - id, sireid, damid, childid

Auth:

Search/view is public access
Data admin is allowed via Okta

Use cases:

  Retrieval:
  - look up female's children
  - look up male's children
  - look up couple's children
  - look up dog's siblings

  Insert/Update:
  - create family (couple and siblings)
  - update dog status

