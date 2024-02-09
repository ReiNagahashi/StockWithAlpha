package models

import (
	"errors"
	"log"

	"gorm.io/gorm"
)

type CompanyFundamental struct {
	Symbol   string `json:"Symbol"`
	Sector   string `json:"Sector"`
	Industry string `json:"Industry"`
	// 株価収益率は企業の収益性を評価するために使用されます 1-3
	PERatio float64 `json:"PERatio"`
	// 成長に関連するP/E比率で、成長を考慮した株価の評価を提供します。
	PEGRatio float64 `json:"PEGRatio"`
	// 年間の四半期収益成長率は企業の成長を示します。1-2
	QuarterlyEarningsGrowthYOY float64 `json:"QuarterlyEarningsGrowthYOY"`
	// 年間の四半期売上成長率も成長を示します。1-3
	QuarterlyRevenueGrowthYOY float64 `json:"QuarterlyRevenueGrowthYOY"`
	// 総資産利益率は資産がどれだけ効率的に使用されているかを示します。1-5
	ReturnOnAssetsTTM float64 `json:"ReturnOnAssetsTTM"`
	// 自己資本利益率は企業が所有者の資本からどれだけ利益を生み出しているかを示します。1-6
	ReturnOnEquityTTM float64 `json:"ReturnOnEquityTTM"`
	// 時価総額は企業の市場での規模とその安定性を示します。
	MarketCapitalization int `json:"MarketCapitalization"`
	// トータルリターンマーケットでの売上高は企業の規模を示します。
	RevenueTTM int `json:"RevenueTTM"`
	// 利払い、税金、減価償却前の収益は企業の運営効率を示します。
	Ebitda int `json:"EBITDA"`
	// 営業利益率は企業の運営効率を示します。1-1
	OperatingMarginTTM float64 `json:"OperatingMarginTTM"`
	// 一株当たり配当金は企業の利益還元を示します。
	DividendPerShare float64 `json:"DividendPerShare"`
	// 配当利回りは投資に対する利益の割合を示します。
	DividendYield float64 `json:"DividendYield"`
	// ベータ値は株式の市場リスクを評価するために使用されます。
	Beta float64 `json:"Beta"`
}

type CompaniesFundamental struct {
	Companies []CompanyFundamental `json:"companies"`
}

func CreateCompanyFundamental(companyFundamental CompanyFundamental) bool {
	tableExists := Goram_db.Migrator().HasTable(&CompanyFundamental{})
	if !tableExists {
		Goram_db.Migrator().CreateTable(&CompanyFundamental{})
	}

	res := Goram_db.Where("symbol = ?", companyFundamental.Symbol).First(&companyFundamental)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		Goram_db.Create(&companyFundamental)
	} else if res.Error != nil {
		log.Fatalln("Error Occured on fetching companyFundamental with symbol")
	} else {
		return false
	}

	return true
}


func GetCompanyBySymbol(symbol string) *CompanyFundamental {
	companyFundamental := CompanyFundamental{}
	res := Goram_db.Where("symbol = ?", symbol).First(&companyFundamental)

	if res.Error != nil {
		log.Println("Failed to get company by industry.")
	}

	return &companyFundamental
}
