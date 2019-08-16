package main

import (
	"encoding/csv"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

var (
	verbose             = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	lumpSums            = kingpin.Arg("lumpSums", "Lump sums file (csv) with format: month-yyyy,amount").Required().File()
	loanAmount          = kingpin.Flag("loanAmount", "Initial loan amount").Required().Float()
	startDate           = kingpin.Flag("startDate", "Start date of loan repayment in format (month yyyy such as September 2019)").Required().String()
	projectedEndBalance = kingpin.Flag("endBalance", "Projected end balance at the end of the term").Default("0.").Float()
	interest            = kingpin.Flag("interest", "Interest rate (i.e. 5 for 5%%)").Required().Float()
	years               = kingpin.Flag("years", "The term in number of years").Required().Int()
)

const (
	version           = "1.0.0"
	paymentTimeFormat = "January 2006"
)

type PaymentTime struct {
	month time.Month
	year  int
}

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	monthlyInterest := *interest / 100.0 / 12.0
	verboseWriter := ioutil.Discard
	if *verbose {
		verboseWriter = os.Stderr
	}
	verboseLog := log.New(verboseWriter, "", log.LstdFlags)

	paymentCount := getPaymentCount(*years)
	monthlyPayment := getMonthlyPayment(monthlyInterest, *loanAmount, paymentCount)
	startDate, err := getMonthYearDate(*startDate)
	if err != nil {
		log.Fatalf("Invalid start date: %s", err.Error())
	}

	verboseLog.Printf("Number of payments is %d, monthly interest rate is %f and monthly payment is %f", paymentCount, monthlyInterest, monthlyPayment)

	lumpSums, err := getLumpSums(verboseLog, *lumpSums)
	if err != nil {
		log.Fatalf("Error reading lump sum files: %s", err.Error())
	}

	balance := *loanAmount
	w.Write([]string{"month", "interest", "principal", "balance"})

	for n := 0; n < paymentCount && balance > 0.0; n++ {
		monthInterest := getInterest(balance, monthlyInterest, n+1)
		monthPrincipal := monthlyPayment - monthInterest
		balance = balance - monthPrincipal
		month := startDate.AddDate(0, n, 0)

		w.Write([]string{fmt.Sprintf("%s", month.Format(paymentTimeFormat)), fmt.Sprintf("%.2f", monthInterest), fmt.Sprintf("%.2f", monthPrincipal), fmt.Sprintf("%.2f", balance)})

		if payment, ok := lumpSums[PaymentTime{month: month.Month(), year: month.Year()}]; ok {
			balance = balance - payment
			w.Write([]string{fmt.Sprintf("%s", month.Format(paymentTimeFormat)), fmt.Sprintf("%.2f", 0.0), fmt.Sprintf("%.2f", payment), fmt.Sprintf("%.2f", balance)})
		}
	}
}

func getLumpSums(verboseLog *log.Logger, lumpSumsFile *os.File) (lumpSums map[PaymentTime]float64, err error) {
	lumpSums = make(map[PaymentTime]float64)

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

		payment, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			if line == 0 {
				verboseLog.Printf("Skipping what looks like a header row: %v", record)
				continue
			} else {
				return nil, errors.Wrapf(err, "Error reading payment value at line %d", line)
			}
		}

		lumpSums[PaymentTime{month: time.Month(), year: time.Year()}] = payment
	}

	return lumpSums, nil
}

func getMonthYearDate(val string) (startDate time.Time, err error) {
	return time.Parse(paymentTimeFormat, val)
}

func getInterest(principal float64, monthlyRate float64, n int) (interest float64) {
	interest = principal * monthlyRate
	return interest
}

func getPaymentCount(term int) (count int) {
	return term * 12
}

func getMonthlyPayment(monthlyRate float64, loanAmount float64, paymentCount int) (monthlyPayment float64) {
	monthlyPayment = monthlyRate / (1.0 - math.Pow(1.0+monthlyRate, float64(paymentCount*-1))) * loanAmount
	return monthlyPayment
}
