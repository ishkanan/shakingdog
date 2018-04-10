DB tables:

Status - id, status
Dog - id, name, gender, shakingdogstatus, cecsstatus
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
  - update shaking dog status
    - cannot change FROM red TO anything else
    - inferred status:
      - clear by parentage: applies to the progeny when both parents test clear
      - carrier by progeny: applies to both parents when any of the progeny test affected
    - inferred status can be changed but display a UI warning
  - update cecs status
    - TBC
