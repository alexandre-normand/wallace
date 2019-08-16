# Wallace


# Combine it with other tools for nicer output

## To get markdown output: 
`Csv` is readable but `markdown` is better. Use [mdt](https://github.com/monochromegane/mdt) 
to convert `csv` to `markdown` tables.

To install `mdt`:
```
go get github.com/monochromegane/mdt/cmd/mdt
```

To combine `wallace` with `mdt`:
```
wallace --loanAmount=125000 --startDate="September 2019" --years=10 --interest 5 --verbose ./lumpSums.csv  | mdt
```

## To get html output:

Use [gomarkdown](https://github.com/gomarkdown/markdown) to convert `markdown` to `html`.

To install `gomarkdown` (the cli):
```
go get github.com/gomarkdown/mdtohtml
```

To combine `wallace` with `mdt` and `gomarkdown`:
```
wallace --loanAmount=125000 --startDate="September 2019" --years=10 --interest 5 --verbose ./lumpSums.csv  | mdt | mdtohtml -css ./github-markdown.css -page > out.html
```