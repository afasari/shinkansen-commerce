You are amazing staff Engineer who implement best practices,

I'm a senior backend software engineer who want to work to japan. I want to create a portfolio.
as my domain knowledge is from e-commerce, and i want to work to japan, so help me create that requirement. I saw saleor and magento is a good example. What do you think, is it better create requirement from scratch or copy and improve from existing project?
as my background in golang, i want use golang as primary, and rust and python as I'm learning both, also i'm open for other stack that japan need

the project name is shinkansen (bullet train japan) that means this ecommerce is fast, reliable, scallable etc And the market is from Japan, so use best practice that big company like rakuten, paypay used

it's a monorepo, with multiple languages to show that i also know that languange
main language is go, some of them will rust for performance, and using pythong for scripting, analytic and low speed services

i want you create the whole flow, this is the draft flow requirements (functional, non functional, hld), create infrastructure & devops, lld & protobuf (contract), implementation, testing, documentation & scale.

The "Japan-Style" Ecommerce 
We have designed a High-Performance, Spec-First Polyglot Monorepo. 
     Philosophy: The Specification (.proto) is the source of truth. Code is just a byproduct.
     The "Japan" Tech Stack:
         Language: Go (Core/Gateway), Python (AI, Simple Services), Rust (Performance).
         Spec/Contract: Protocol Buffers managed by Buf (instead of raw OpenAPI/YAML).
         Data Access: sqlc (SQL -> Code generation) instead of ORMs.
         Communication: gRPC (Internal), REST (External).
         Database: Postgres (Logical separation via Schemas).
         
     The Goal: Prove you can build a system that is decoupled, type-safe, and ready for scale ("Million Users" mindset).
