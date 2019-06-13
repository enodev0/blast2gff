#blast2gff

Convert blastn alignment files to GFFv3 annotation format, for example after aligning de-novo predicted mRNA transcripts to a reference genome.
In case of strand specific alignment, it can also merge the corresponding GFFs.

## Usage
```bash
~$ blast2gff convert watson alignment_fowward_strand.aln > watson.gff
```
or perhaps ...
```bash
~$ blast2gff convert crick alignment_reverse_strand.aln > crick.gff
```
we can then merge as follows
```bash
~$ blast2gff merge watson.gff crick.gff > transcript_annotation.gff
```

## License
MIT
