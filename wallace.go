package main

import (
	"encoding/csv"
	"fmt"
	"github.com/ALTree/floatutils"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/leekchan/accounting"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"
	"time"
)

var (
	verbose             = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	lumpSums            = kingpin.Arg("lumpSums", "Lump sums file (csv) with format: month-yyyy,amount").Required().File()
	loanAmount          = kingpin.Flag("loanAmount", "Initial loan amount").Required().Float()
	startDate           = kingpin.Flag("startDate", "Start date of loan repayment in format (month yyyy such as September 9 2019)").Required().String()
	projectedEndBalance = kingpin.Flag("endBalance", "Projected end balance at the end of the term").Default("0.").Float()
	interest            = kingpin.Flag("interest", "Interest rate (i.e. 5 for 5%%)").Required().Float()
	years               = kingpin.Flag("years", "The term in number of years").Required().Int()
	compounding         = kingpin.Flag("compounding", "The interest compounding").Default("semiannually").Enum("semiannually")
	output              = kingpin.Flag("output", "The output format").Default("csv").Enum("csv", "markdown", "html")
)

const (
	version           = "1.0.0"
	paymentTimeFormat = "January 2 2006"
	cssContent        = "@font-face{font-family:octicons-link;src:url(data:font/woff;charset=utf-8;base64,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) format('woff')}body{box-sizing:border-box;min-width:200px;max-width:980px;margin:0 auto;padding:45px;-ms-text-size-adjust:100%;-webkit-text-size-adjust:100%;color:#24292e;line-height:1.5;font-family:-apple-system,BlinkMacSystemFont,Segoe UI,Helvetica,Arial,sans-serif,Apple Color Emoji,Segoe UI Emoji,Segoe UI Symbol;font-size:16px;line-height:1.5;word-wrap:break-word}table{border-collapse:collapse;border-spacing:0}td,th{padding:0}table{margin-bottom:16px;margin-top:0;display:block;overflow:auto;width:100%}table th{font-weight:600}table td,table th{border:1px solid #dfe2e5;padding:6px 13px}table tr{background-color:#fff;border-top:1px solid #c6cbd1}table tr:nth-child(2n){background-color:#f6f8fa}"
)

var currency = accounting.Accounting{Symbol: "$", Thousand: ",", Precision: 2}
var bigZero = big.NewFloat(0.0)

type PaymentTime struct {
	month time.Month
	year  int
}

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	verboseWriter := ioutil.Discard
	if *verbose {
		verboseWriter = os.Stderr
	}
	verboseLog := log.New(verboseWriter, "", log.LstdFlags)

	annualInterest := big.NewFloat(0.0).Quo(big.NewFloat(*interest), big.NewFloat(100.0))
	verboseLog.Printf("Annual interest: %.2f%%", annualInterest)
	monthlyInterest := annualInterest.Quo(annualInterest, big.NewFloat(12.0))
	verboseLog.Printf("monthly interest: %f%%", monthlyInterest)

	paymentCount := getPaymentCount(*years)
	verboseLog.Printf("Number of payments is %d, monthly interest rate is %f%%", paymentCount, monthlyInterest)
	monthlyPayment := getMonthlyPayment(*monthlyInterest, *big.NewFloat(*loanAmount), paymentCount)
	startDate, err := getMonthYearDate(*startDate)
	if err != nil {
		log.Fatalf("Invalid start date: %s", err.Error())
	}

	verboseLog.Printf("Output mode: %s", *output)
	verboseLog.Printf("Number of payments is %d, monthly interest rate is %f%% and monthly payment is %s", paymentCount, monthlyInterest, currency.FormatMoneyBigFloat(&monthlyPayment))

	lumpSums, err := getLumpSums(verboseLog, *lumpSums)
	if err != nil {
		log.Fatalf("Error reading lump sum files: %s", err.Error())
	}

	balance := big.NewFloat(*loanAmount)

	var csvBuilder strings.Builder
	w := csv.NewWriter(&csvBuilder)
	w.Write([]string{"month", "interest", "principal", "balance"})

	for n := 0; n < paymentCount && balance.Cmp(bigZero) > 0; n++ {
		monthInterest := getInterest(*balance, *monthlyInterest, n+1)
		monthlyPayment = bigFloatMin(monthlyPayment, *big.NewFloat(0.0).Add(balance, &monthInterest))
		monthPrincipal := big.NewFloat(0.0).Sub(&monthlyPayment, &monthInterest)
		balance = balance.Sub(balance, monthPrincipal)
		month := startDate.AddDate(0, n, 0)

		w.Write([]string{fmt.Sprintf("%s", month.Format(paymentTimeFormat)), fmt.Sprintf("%s", currency.FormatMoneyBigFloat(&monthInterest)), fmt.Sprintf("%s", currency.FormatMoneyBigFloat(monthPrincipal)), fmt.Sprintf("%s", currency.FormatMoneyBigFloat(balance))})

		if payment, ok := lumpSums[PaymentTime{month: month.Month(), year: month.Year()}]; ok {
			balance = balance.Sub(balance, &payment)
			w.Write([]string{fmt.Sprintf("%s", month.Format(paymentTimeFormat)), "none", fmt.Sprintf("%s", currency.FormatMoneyBigFloat(&payment)), fmt.Sprintf("%s", currency.FormatMoneyBigFloat(balance))})
		}
	}
	w.Flush()

	// If we're outputting csv, stop here and dump the output
	if *output == "csv" {
		fmt.Fprint(os.Stdout, csvBuilder.String())
		return
	}

	mrkdwn, err := csvToMarkdown(csvBuilder.String())
	if err != nil {
		log.Fatalf("Error rendering markdown: %s", err.Error())
	}

	if *output == "markdown" {
		fmt.Fprint(os.Stdout, mrkdwn)
		return
	} else {
		opts := html.RendererOptions{Flags: html.CommonFlags | html.CompletePage,
			Title: "Wallace Report"}
		renderer := html.NewRenderer(opts)

		html := markdown.ToHTML([]byte(mrkdwn), nil, renderer)
		// Insert css in the html header
		styledHTML := strings.Replace(string(html), "</head>", fmt.Sprintf("   <style>\n%s\n   </style>\n</head>", cssContent), 1)

		fmt.Fprintf(os.Stdout, styledHTML)
	}
}

func csvToMarkdown(rawCsv string) (mrkdwn string, err error) {
	r := csv.NewReader(strings.NewReader(rawCsv))

	records, err := r.ReadAll()
	if err != nil {
		return "", err
	}

	var b strings.Builder
	table := tablewriter.NewWriter(&b)
	table.SetHeader(records[0])
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(records[1:])
	table.Render()

	return b.String(), nil
}

func bigFloatMin(a big.Float, b big.Float) (min big.Float) {
	if a.Cmp(&b) < 0 {
		return a
	} else {
		return b
	}
}

func getLumpSums(verboseLog *log.Logger, lumpSumsFile *os.File) (lumpSums map[PaymentTime]big.Float, err error) {
	lumpSums = make(map[PaymentTime]big.Float)

	r := csv.NewReader(lumpSumsFile)

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	for line, record := range records {
		if len(record) != 2 {
			return nil, fmt.Errorf("Incorrect format, should be: paymentTime,paymentValue but was %v", record)
		}

		time, err := getMonthYearDate(record[0])
		if err != nil {
			if line == 0 {
				verboseLog.Printf("Skipping what looks like a header row: %v", record)
				continue
			} else {
				return nil, errors.Wrapf(err, "Error reading payment time at line %d, should be in format month yyyy (i.e. January 2006)", line)
			}
		}

		payment, _, err := big.ParseFloat(record[1], 10, 53, big.ToNearestEven)
		if err != nil {
			if line == 0 {
				verboseLog.Printf("Skipping what looks like a header row: %v", record)
				continue
			} else {
				return nil, errors.Wrapf(err, "Error reading payment value at line %d", line)
			}
		}

		lumpSums[PaymentTime{month: time.Month(), year: time.Year()}] = *payment
	}

	return lumpSums, nil
}

func getMonthYearDate(val string) (startDate time.Time, err error) {
	return time.Parse(paymentTimeFormat, val)
}

func getInterest(principal big.Float, monthlyRate big.Float, n int) (interest big.Float) {
	interest = *big.NewFloat(0.0).Mul(&principal, &monthlyRate)
	return interest
}

func getPaymentCount(term int) (count int) {
	return term * 12
}

func getMonthlyPayment(monthlyRate big.Float, loanAmount big.Float, paymentCount int) (monthlyPayment big.Float) {
	// 1.0+monthlyRate
	powA := big.NewFloat(0.0).Add(big.NewFloat(1.0), &monthlyRate)
	// float64(paymentCount*-1)
	pc := big.NewFloat(float64(paymentCount))
	negPaymentCount := pc.Neg(pc)
	// math.Pow(1.0+monthlyRate, float64(paymentCount*-1))
	pow := bigfloat.Pow(powA, negPaymentCount)
	// (1.0 - math.Pow(1.0+monthlyRate, float64(paymentCount*-1))
	divisor := big.NewFloat(0.0).Sub(big.NewFloat(1.0), pow)
	// monthlyRate / (1.0 - math.Pow(1.0+monthlyRate, float64(paymentCount*-1)))
	monthlyPay := big.NewFloat(0.0).Quo(&monthlyRate, divisor)
	// monthlyPayment = monthlyRate / (1.0 - math.Pow(1.0+monthlyRate, float64(paymentCount*-1))) * loanAmount
	monthlyPay = monthlyPayment.Mul(monthlyPay, &loanAmount)
	return *monthlyPay
}
