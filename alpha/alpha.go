package alpha

// やっと理解した。Tickerの構造体がここにある理由は、APIから直接持ってくる構造体
// →APIと通信するためにはAPIClientが必要になる。
// 一方で、candleはTikerの特定のフィールドを引数として渡して自身で作った構造体
// その上で、dbでの処理をするので、modelにcandeが置いてあるのだ
// さもなければ、例えばORMを組もうとしても、どこのモデルとモデルを関連付けさせるのかが分からなくなる
// つまり、alpha antage以外のAPIを使う場合は、新たにフォルダを作成する

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"stock-with-alpha/app/models"
	"stock-with-alpha/config"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://www.alphavantage.co/"

type APIClient struct{
	key 		string
	httpClient 	*http.Client
}


func New(key string) *APIClient{
	apiClient := &APIClient{key, &http.Client{}}
	return apiClient
}


// リクエストをwebサーバに投げる
func (api *APIClient) doRequest(method, urlPath string, query map[string]string, data []byte) (body []byte, err error){
	baseURL, err := url.Parse(baseURL)
	if err != nil{
		return 
	}

	urlValues := url.Values{}
	for key, val := range query{
		urlValues.Add(key, val)
	}
	
	if len(query) > 0{
		urlPath = urlPath + "query?" + urlValues.Encode()
	}

	apiURL, err := url.Parse(urlPath)
	if err != nil{
		return
	}

	endpoint := baseURL.ResolveReference(apiURL).String()
	fmt.Printf("action=doRequest, endpoint=%s", endpoint)

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(data))
	if err != nil{
		return
	}
	req.URL.RawQuery = req.URL.Query().Encode()

	resp, err := api.httpClient.Do(req)
	if err != nil{
		return nil, err
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil{
		return nil, err
	}

	return body, nil
}

// Meta data for Weekly
type MetaDataWeekly struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	TimeZone      string `json:"4. Time Zone"`
	
}

// Meta data for Daily
type MetaDataDaily struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize	string 	 `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}


// TimeSeriesEntry はAPIから取得したデータの1エントリを表す構造体です。
type TimeSeriesEntry struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5. volume"`
}

type Ticker struct {
	Symbol string
	Date  string          // 日付を格納するフィールド
	Data  TimeSeriesEntry // 日付に対応するデータを格納するフィールド
}

// Response はAPIのレスポンス全体を表す構造体です。
type ResponseWeekly struct {
	MetaDataWeekly         MetaDataWeekly         			`json:"Meta Data"`
	Ticker 			 map[string]TimeSeriesEntry `json:"Weekly Time Series"`
}

// Response Daily
type ResponseDaily struct {
	MetaDataDaily         MetaDataDaily         			`json:"Meta Data"`
	Ticker 			 map[string]TimeSeriesEntry `json:"Time Series (Daily)"`
}

func (api *APIClient) GetDailyTicker(symbol, name, f string, duration time.Duration, ch chan <- error) {
	resp, err := api.doRequest("GET", "", map[string]string{"symbol": symbol, "function": f, "apikey": config.Config.ApiKey}, nil)
	if err != nil{
		ch <- err
	}
	
	var response = ResponseDaily{}
	
	if err := json.Unmarshal([]byte(resp), &response); err != nil {
		ch <- err
	}

	for date, data := range response.Ticker {
		symbol := response.MetaDataDaily.Symbol
		name := name
		open,_ := strconv.ParseFloat(data.Open, 64)
		close,_ := strconv.ParseFloat(data.Close, 64)
		high, _ := strconv.ParseFloat(data.High, 64)
		low, _ := strconv.ParseFloat(data.Low, 64)
		volume, _ := strconv.ParseFloat(data.Volume, 64)
		datePart := strings.Split(date, " ")[0]

		parsedDate, err := time.Parse("2006-01-02", datePart)
		if err != nil {
			ch <- err
		}
		
		candle := models.Candle{
			Symbol: symbol,
			Name: name,
			Duration: duration,
			Open: open,
			Close: close,
			High: high,
			Low: low,
			Volume: volume,
			DateTime: parsedDate,
		}

		models.CreateCandleWithDuration(candle, symbol, name, parsedDate, duration)

	}
}



// func (api *APIClient) GetWeeklyTicker(symbol, f string, duration time.Duration) error{
// 	resp, err := api.doRequest("GET", "", map[string]string{"symbol": symbol, "function": f, "apikey": config.Config.ApiKey}, nil)
// 	if err != nil{
// 		return err
// 	}
	
// 	var response = ResponseWeekly{}
	
// 	if err := json.Unmarshal([]byte(resp), &response); err != nil {
// 		return err
// 	}

// 	for date, data := range response.Ticker {
// 		symbol := response.MetaDataWeekly.Symbol
// 		open,_ := strconv.ParseFloat(data.Open, 64)
// 		close,_ := strconv.ParseFloat(data.Close, 64)
// 		high, _ := strconv.ParseFloat(data.High, 64)
// 		low, _ := strconv.ParseFloat(data.Low, 64)
// 		volume, _ := strconv.ParseFloat(data.Volume, 64)
// 		datePart := strings.Split(date, " ")[0]

// 		parsedDate, err := time.Parse("2006-01-02", datePart)
// 		if err != nil {
// 			log.Println("Error parsing date:", err)
// 		}

// 		candle := models.Candle{
// 			Symbol: symbol,
// 			Duration: duration,
// 			Open: open,
// 			Close: close,
// 			High: high,
// 			Low: low,
// 			Volume: volume,
// 			DateTime: parsedDate,
// 		}


// 		models.CreateCandleWithDuration(candle, symbol, name, parsedDate, duration)

// 	}
	
// 	return nil
// }


type TickerByKeyword struct {
	BestMatches []struct {
		Symbol      string `json:"1. symbol"`
		Name        string `json:"2. name"`
		Type      string `json:"3. type"`
		Region     string `json:"4. region"`
		MarketOpen string `json:"5. marketOpen"`
		MarketClose string `json:"6. marketClose"`
		Timezone  string `json:"7. timezone"`
		Currency  string `json:"8. currency"`
		MatchScore string `json:"9. matchScore"`
	} `json:"bestMatches"`
}


func (api *APIClient) GetTickerInfo(f, keyword string) (*TickerByKeyword, error){
	resp, err := api.doRequest("GET", "", map[string]string{"function": f, "keywords": keyword, "apikey": config.Config.ApiKey}, nil)
	if err != nil{
		return nil, err
	}
	
	var response = TickerByKeyword{}

	if err := json.Unmarshal([]byte(resp), &response); err != nil {
		return nil, err
	}

	return &response, nil

}

// Company Overview
type CompanyOverview struct {
	Symbol                     string `json:"Symbol"`
	AssetType                  string `json:"AssetType"`
	Name                       string `json:"Name"`
	Description                string `json:"Description"`
	Cik                        string `json:"CIK"`
	Exchange                   string `json:"Exchange"`
	Currency                   string `json:"Currency"`
	Country                    string `json:"Country"`
	Sector                     string `json:"Sector"`
	Industry                   string `json:"Industry"`
	Address                    string `json:"Address"`
	FiscalYearEnd              string `json:"FiscalYearEnd"`
	LatestQuarter              string `json:"LatestQuarter"`
	MarketCapitalization       string `json:"MarketCapitalization"`
	Ebitda                     string `json:"EBITDA"`
	PERatio                    string `json:"PERatio"`
	PEGRatio                   string `json:"PEGRatio"`
	BookValue                  string `json:"BookValue"`
	DividendPerShare           string `json:"DividendPerShare"`
	DividendYield              string `json:"DividendYield"`
	Eps                        string `json:"EPS"`
	RevenuePerShareTTM         string `json:"RevenuePerShareTTM"`
	ProfitMargin               string `json:"ProfitMargin"`
	OperatingMarginTTM         string `json:"OperatingMarginTTM"`
	ReturnOnAssetsTTM          string `json:"ReturnOnAssetsTTM"`
	ReturnOnEquityTTM          string `json:"ReturnOnEquityTTM"`
	RevenueTTM                 string `json:"RevenueTTM"`
	GrossProfitTTM             string `json:"GrossProfitTTM"`
	DilutedEPSTTM              string `json:"DilutedEPSTTM"`
	QuarterlyEarningsGrowthYOY string `json:"QuarterlyEarningsGrowthYOY"`
	QuarterlyRevenueGrowthYOY  string `json:"QuarterlyRevenueGrowthYOY"`
	AnalystTargetPrice         string `json:"AnalystTargetPrice"`
	TrailingPE                 string `json:"TrailingPE"`
	ForwardPE                  string `json:"ForwardPE"`
	PriceToSalesRatioTTM       string `json:"PriceToSalesRatioTTM"`
	PriceToBookRatio           string `json:"PriceToBookRatio"`
	EVToRevenue                string `json:"EVToRevenue"`
	EVToEBITDA                 string `json:"EVToEBITDA"`
	Beta                       string `json:"Beta"`
	Five2WeekHigh              string `json:"52WeekHigh"`
	Five2WeekLow               string `json:"52WeekLow"`
	Five0DayMovingAverage      string `json:"50DayMovingAverage"`
	Two00DayMovingAverage      string `json:"200DayMovingAverage"`
	SharesOutstanding          string `json:"SharesOutstanding"`
	DividendDate               string `json:"DividendDate"`
	ExDividendDate             string `json:"ExDividendDate"`
}


func (api *APIClient) GetCompanyOverview(symbol string, ch chan <- error) {
	resp, err := api.doRequest("GET", "", map[string]string{"function": "OVERVIEW", "symbol": symbol, "apikey": config.Config.ApiKey}, nil)
	if err != nil{
		ch <- err
	}
	
	var response = CompanyOverview{}

	if err := json.Unmarshal([]byte(resp), &response); err != nil {
		ch <- err
	}

	pegRatio,_ := strconv.ParseFloat(response.PEGRatio, 64)
	pEGRatio,_ := strconv.ParseFloat(response.PEGRatio, 64)
	quarterlyEarningsGrowthYOY,_ := strconv.ParseFloat(response.QuarterlyEarningsGrowthYOY, 64)
	quarterlyRevenueGrowthYOY,_ := strconv.ParseFloat(response.QuarterlyRevenueGrowthYOY, 64)
	returnOnAssetsTTM,_ := strconv.ParseFloat(response.ReturnOnAssetsTTM, 64)
	returnOnEquityTTM,_ := strconv.ParseFloat(response.ReturnOnEquityTTM, 64)
	operatingMarginTTM,_ := strconv.ParseFloat(response.OperatingMarginTTM, 64)
	dividendPerShare,_ := strconv.ParseFloat(response.DividendPerShare, 64)
	dividendYield,_ := strconv.ParseFloat(response.DividendYield, 64)
	beta,_ := strconv.ParseFloat(response.Beta, 64)
	marketCapitalization,_ := strconv.Atoi(response.MarketCapitalization)
	revenueTTM,_ := strconv.Atoi(response.RevenueTTM)
	ebitda,_ := strconv.Atoi(response.Ebitda)

	
	companyFundamental := models.CompanyFundamental{
		Symbol: response.Symbol,
		Sector: response.Sector,
		Industry: response.Industry,
		PERatio: pegRatio,
		PEGRatio: pEGRatio,
		QuarterlyEarningsGrowthYOY: quarterlyEarningsGrowthYOY,
		QuarterlyRevenueGrowthYOY: quarterlyRevenueGrowthYOY,
		ReturnOnAssetsTTM: returnOnAssetsTTM,
		ReturnOnEquityTTM: returnOnEquityTTM,
		OperatingMarginTTM: operatingMarginTTM,
		DividendPerShare: dividendPerShare,
		DividendYield: dividendYield,
		Beta: beta,
		MarketCapitalization: marketCapitalization,
		RevenueTTM: revenueTTM,
		Ebitda: ebitda,

	}

	if (models.CreateCompanyFundamental(companyFundamental)){
		fmt.Println("CompanyFundamental data inserted!")
	}

}


type BSAnnualReport struct{
	FiscalDateEnding                       string `json:"fiscalDateEnding"`
	ReportedCurrency                       string `json:"reportedCurrency"`
	TotalAssets                            string `json:"totalAssets"`
	TotalCurrentAssets                     string `json:"totalCurrentAssets"`
	CashAndCashEquivalentsAtCarryingValue  string `json:"cashAndCashEquivalentsAtCarryingValue"`
	CashAndShortTermInvestments            string `json:"cashAndShortTermInvestments"`
	Inventory                              string `json:"inventory"`
	CurrentNetReceivables                  string `json:"currentNetReceivables"`
	TotalNonCurrentAssets                  string `json:"totalNonCurrentAssets"`
	PropertyPlantEquipment                 string `json:"propertyPlantEquipment"`
	AccumulatedDepreciationAmortizationPPE string `json:"accumulatedDepreciationAmortizationPPE"`
	IntangibleAssets                       string `json:"intangibleAssets"`
	IntangibleAssetsExcludingGoodwill      string `json:"intangibleAssetsExcludingGoodwill"`
	Goodwill                               string `json:"goodwill"`
	Investments                            string `json:"investments"`
	LongTermInvestments                    string `json:"longTermInvestments"`
	ShortTermInvestments                   string `json:"shortTermInvestments"`
	OtherCurrentAssets                     string `json:"otherCurrentAssets"`
	OtherNonCurrentAssets                  string `json:"otherNonCurrentAssets"`
	TotalLiabilities                       string `json:"totalLiabilities"`
	TotalCurrentLiabilities                string `json:"totalCurrentLiabilities"`
	CurrentAccountsPayable                 string `json:"currentAccountsPayable"`
	DeferredRevenue                        string `json:"deferredRevenue"`
	CurrentDebt                            string `json:"currentDebt"`
	ShortTermDebt                          string `json:"shortTermDebt"`
	TotalNonCurrentLiabilities             string `json:"totalNonCurrentLiabilities"`
	CapitalLeaseObligations                string `json:"capitalLeaseObligations"`
	LongTermDebt                           string `json:"longTermDebt"`
	CurrentLongTermDebt                    string `json:"currentLongTermDebt"`
	LongTermDebtNoncurrent                 string `json:"longTermDebtNoncurrent"`
	ShortLongTermDebtTotal                 string `json:"shortLongTermDebtTotal"`
	OtherCurrentLiabilities                string `json:"otherCurrentLiabilities"`
	OtherNonCurrentLiabilities             string `json:"otherNonCurrentLiabilities"`
	TotalShareholderEquity                 string `json:"totalShareholderEquity"`
	TreasuryStock                          string `json:"treasuryStock"`
	RetainedEarnings                       string `json:"retainedEarnings"`
	CommonStock                            string `json:"commonStock"`
	CommonStockSharesOutstanding           string `json:"commonStockSharesOutstanding"`
}

//AnnualReportsの一部のみを取りたいので、AnnualReportという構造体を作って扱いやすいようにした
type BS struct {
	Symbol        string 			`json:"symbol"`
	AnnualReports []BSAnnualReport 	`json:"annualReports"`
}


func (api *APIClient) GetBS(symbol string) []BSAnnualReport {
	resp, err := api.doRequest("GET", "", map[string]string{"function": "BALANCE_SHEET", "symbol": symbol, "apikey": config.Config.ApiKey}, nil)
	if err != nil{
		log.Fatalln(err)
		return nil
	}
	
	var response = BS{}

	if err := json.Unmarshal([]byte(resp), &response); err != nil {
		log.Fatalln(err)
		return nil
	}

	return response.AnnualReports[:2]

}


type PLAnnualReport struct {
	FiscalDateEnding                  string `json:"fiscalDateEnding"`
	ReportedCurrency                  string `json:"reportedCurrency"`
	GrossProfit                       string `json:"grossProfit"`
	TotalRevenue                      string `json:"totalRevenue"`
	CostOfRevenue                     string `json:"costOfRevenue"`
	CostofGoodsAndServicesSold        string `json:"costofGoodsAndServicesSold"`
	OperatingIncome                   string `json:"operatingIncome"`
	SellingGeneralAndAdministrative   string `json:"sellingGeneralAndAdministrative"`
	ResearchAndDevelopment            string `json:"researchAndDevelopment"`
	OperatingExpenses                 string `json:"operatingExpenses"`
	InvestmentIncomeNet               string `json:"investmentIncomeNet"`
	NetInterestIncome                 string `json:"netInterestIncome"`
	InterestIncome                    string `json:"interestIncome"`
	InterestExpense                   string `json:"interestExpense"`
	NonInterestIncome                 string `json:"nonInterestIncome"`
	OtherNonOperatingIncome           string `json:"otherNonOperatingIncome"`
	Depreciation                      string `json:"depreciation"`
	DepreciationAndAmortization       string `json:"depreciationAndAmortization"`
	IncomeBeforeTax                   string `json:"incomeBeforeTax"`
	IncomeTaxExpense                  string `json:"incomeTaxExpense"`
	InterestAndDebtExpense            string `json:"interestAndDebtExpense"`
	NetIncomeFromContinuingOperations string `json:"netIncomeFromContinuingOperations"`
	ComprehensiveIncomeNetOfTax       string `json:"comprehensiveIncomeNetOfTax"`
	Ebit                              string `json:"ebit"`
	Ebitda                            string `json:"ebitda"`
	NetIncome                         string `json:"netIncome"`
}


//AnnualReportsの一部のみを取りたいので、AnnualReportという構造体を作って扱いやすいようにした
type PL struct {
	Symbol        string 			`json:"symbol"`
	AnnualReports []PLAnnualReport 	`json:"annualReports"`
}


func (api *APIClient) GetPL(symbol string) []PLAnnualReport {
	resp, err := api.doRequest("GET", "", map[string]string{"function": "INCOME_STATEMENT", "symbol": symbol, "apikey": config.Config.ApiKey}, nil)
	if err != nil{
		log.Fatalln(err)
		return nil
	}
	
	var response = PL{}
	if err := json.Unmarshal([]byte(resp), &response); err != nil {
		log.Fatalln(err)
		return nil
	}

	return response.AnnualReports[:2]

}


type CFAnnualReport struct {
	FiscalDateEnding                                          string `json:"fiscalDateEnding"`
	ReportedCurrency                                          string `json:"reportedCurrency"`
	OperatingCashflow                                         string `json:"operatingCashflow"`
	PaymentsForOperatingActivities                            string `json:"paymentsForOperatingActivities"`
	ProceedsFromOperatingActivities                           string `json:"proceedsFromOperatingActivities"`
	ChangeInOperatingLiabilities                              string `json:"changeInOperatingLiabilities"`
	ChangeInOperatingAssets                                   string `json:"changeInOperatingAssets"`
	DepreciationDepletionAndAmortization                      string `json:"depreciationDepletionAndAmortization"`
	CapitalExpenditures                                       string `json:"capitalExpenditures"`
	ChangeInReceivables                                       string `json:"changeInReceivables"`
	ChangeInInventory                                         string `json:"changeInInventory"`
	ProfitLoss                                                string `json:"profitLoss"`
	CashflowFromInvestment                                    string `json:"cashflowFromInvestment"`
	CashflowFromFinancing                                     string `json:"cashflowFromFinancing"`
	ProceedsFromRepaymentsOfShortTermDebt                     string `json:"proceedsFromRepaymentsOfShortTermDebt"`
	PaymentsForRepurchaseOfCommonStock                        string `json:"paymentsForRepurchaseOfCommonStock"`
	PaymentsForRepurchaseOfEquity                             string `json:"paymentsForRepurchaseOfEquity"`
	PaymentsForRepurchaseOfPreferredStock                     string `json:"paymentsForRepurchaseOfPreferredStock"`
	DividendPayout                                            string `json:"dividendPayout"`
	DividendPayoutCommonStock                                 string `json:"dividendPayoutCommonStock"`
	DividendPayoutPreferredStock                              string `json:"dividendPayoutPreferredStock"`
	ProceedsFromIssuanceOfCommonStock                         string `json:"proceedsFromIssuanceOfCommonStock"`
	ProceedsFromIssuanceOfLongTermDebtAndCapitalSecuritiesNet string `json:"proceedsFromIssuanceOfLongTermDebtAndCapitalSecuritiesNet"`
	ProceedsFromIssuanceOfPreferredStock                      string `json:"proceedsFromIssuanceOfPreferredStock"`
	ProceedsFromRepurchaseOfEquity                            string `json:"proceedsFromRepurchaseOfEquity"`
	ProceedsFromSaleOfTreasuryStock                           string `json:"proceedsFromSaleOfTreasuryStock"`
	ChangeInCashAndCashEquivalents                            string `json:"changeInCashAndCashEquivalents"`
	ChangeInExchangeRate                                      string `json:"changeInExchangeRate"`
	NetIncome                                                 string `json:"netIncome"`
}


//AnnualReportsの一部のみを取りたいので、AnnualReportという構造体を作って扱いやすいようにした
type CF struct {
	Symbol        string 			`json:"symbol"`
	AnnualReports []CFAnnualReport 	`json:"annualReports"`
}


func (api *APIClient) GetCF(symbol string) []CFAnnualReport {
	resp, err := api.doRequest("GET", "", map[string]string{"function": "CASH_FLOW", "symbol": symbol, "apikey": config.Config.ApiKey}, nil)
	if err != nil{
		log.Fatalln(err)
		return nil
	}
	
	var response = CF{}

	if err := json.Unmarshal([]byte(resp), &response); err != nil {
		log.Fatalln(err)
		return nil
	}

	return response.AnnualReports[:2]

}
