# Entity normalization

## Canonical-name rule

Use a stable lowercase kebab-case canonical name and a proper display name

Examples:

- `ge-vernova` / `GE Vernova`
- `siemens-energy` / `Siemens Energy`
- `azad-engineering` / `Azad Engineering`

## Reuse before creating

If the case entity registry exists, reuse its canonical name when an alias clearly refers to the same entity. Record the observed variant in `aliases`

## Do not over-merge

Keep separate:

- parent company and operating subsidiary when the distinction matters
- OEM and joint venture
- current company and former legal entity when contracts or evidence differ
- product brand and corporate entity
- similarly named companies without clear identity evidence

## Role is relational

Do not encode `supplier`, `customer`, or `competitor` as the entity's permanent type. Store those as roles and relationships scoped to the evidence

Use `entity_type` for stable categories such as `company`, `government`, `customer-segment`, `product`, `technology`, `geography`, or `market`
