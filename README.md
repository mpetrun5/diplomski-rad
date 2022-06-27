# github.com/mpetrun5/diplomski-rad

Blockchain novčanik za generaciju ECDSA potpisa pomoću kriptografije praga prema radu Gennara i Goldfedera.

Podržava:
  - generaciju ključa
  - osvježavanje ključa
  - slanje transakcija na Ethereum mreži

# CLI

Sva funkcionalost dostupna je preko komandolinijskog sučelja.
Prema funkcionalnosti, podržane su dvije vrste čvora u mreži. Pasivni čvor,
koji čeka na poruku inicijacije TSS procesa te čvor koji započinje TSS proces.

### Pasivni čvor

Pasivni čvor se pokreće pomoću komande:

```bash
./diplomski-rad listen --config <put-do-datoteke>
```

### Generacija ključeva

Nakon što su pasivni čvorovi spremni u mreži, pokretanje procesa generacije ključeva
izvršava se pomoću komande:

```bash
./diplomski-rad generate-key --config <put-do-datoteke>
```

### Osvježavanje ključeva

```bash
./diplomski-rad refresh-key --config <put-do-datoteke>
```

### Slanje transakcije

Slanje transkacije započinje TSS generaciju potpisa te nakon uspješne generacije,
šalje transakciju na čvor Ethereum mreže.

```bash
./diplomski-rad send-transaction --to <Ethereum adresa> --network <RPC URL Ethereum mreže> --data <arbitratni podatci transakcije> --config <put-do-datoteke>
```
