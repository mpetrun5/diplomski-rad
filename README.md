# diplomski

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
./diplomski listen --config <put-do-datoteke>
```

### Generacija ključeva

Nakon što su pasivni čvorovi spremni u mreži, pokretanje procesa generacije ključeva
izvršava se pomoću komande:

```bash
./diplomski generate-key --config <put-do-datoteke>
```

### Osvježavanje ključeva

```bash
./diplomski refresh-key --config <put-do-datoteke>
```
