package dto

import (
	"time"

	"github.com/dujiao-next/internal/models"
)

type ResellerProfileSummaryResp struct {
	ID               uint      `json:"id"`
	Status           string    `json:"status"`
	SettlementStatus string    `json:"settlement_status"`
	CreatedAt        time.Time `json:"created_at"`
}

type ResellerBalanceResp struct {
	ID              uint      `json:"id"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	AvailableAmount string    `json:"available_amount"`
	LockedAmount    string    `json:"locked_amount"`
	NegativeAmount  string    `json:"negative_amount"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ResellerLedgerResp struct {
	ID                uint       `json:"id"`
	OrderID           *uint      `json:"order_id,omitempty"`
	Type              string     `json:"type"`
	Amount            string     `json:"amount"`
	Currency          string     `json:"currency"`
	Status            string     `json:"status"`
	AvailableAt       *time.Time `json:"available_at,omitempty"`
	WithdrawRequestID *uint      `json:"withdraw_request_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type ResellerWithdrawResp struct {
	ID           uint       `json:"id"`
	Amount       string     `json:"amount"`
	Currency     string     `json:"currency"`
	Channel      string     `json:"channel"`
	Account      string     `json:"account"`
	Status       string     `json:"status"`
	RejectReason string     `json:"reject_reason,omitempty"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type ResellerDashboardResp struct {
	Opened                 bool                        `json:"opened"`
	Profile                *ResellerProfileSummaryResp `json:"profile,omitempty"`
	Balances               []ResellerBalanceResp       `json:"balances,omitempty"`
	WithdrawEnabled        bool                        `json:"withdraw_enabled"`
	WithdrawDisabledReason string                      `json:"withdraw_disabled_reason,omitempty"`
}

type ResellerManagementProfileResp struct {
	ID                   uint       `json:"id"`
	Status               string     `json:"status"`
	ApplyReason          string     `json:"apply_reason,omitempty"`
	RejectReason         string     `json:"reject_reason,omitempty"`
	DefaultMarkupPercent string     `json:"default_markup_percent"`
	MaxMarkupPercent     string     `json:"max_markup_percent"`
	SettlementStatus     string     `json:"settlement_status"`
	ReviewedAt           *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type ResellerDomainResp struct {
	ID                 uint       `json:"id"`
	Domain             string     `json:"domain"`
	Type               string     `json:"type"`
	VerificationToken  string     `json:"verification_token,omitempty"`
	VerificationStatus string     `json:"verification_status"`
	Status             string     `json:"status"`
	IsPrimary          bool       `json:"is_primary"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type ResellerManagementSnapshotResp struct {
	Opened   bool                           `json:"opened"`
	CanApply bool                           `json:"can_apply"`
	Profile  *ResellerManagementProfileResp `json:"profile,omitempty"`
	Domains  []ResellerDomainResp           `json:"domains"`
}

type ResellerSiteConfigResp struct {
	ID           uint          `json:"id"`
	SiteName     string        `json:"site_name"`
	Logo         string        `json:"logo"`
	Favicon      string        `json:"favicon"`
	Announcement models.JSON   `json:"announcement"`
	Support      models.JSON   `json:"support"`
	SEO          models.JSON   `json:"seo"`
	FooterLinks  []interface{} `json:"footer_links"`
	NavConfig    models.JSON   `json:"nav_config"`
	Theme        models.JSON   `json:"theme"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type ResellerSiteConfigSnapshotResp struct {
	Opened  bool                    `json:"opened"`
	CanEdit bool                    `json:"can_edit"`
	Config  *ResellerSiteConfigResp `json:"config,omitempty"`
}

type ResellerSiteConfigOwnerUserResp struct {
	ID          uint   `json:"id"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

type ResellerSiteConfigProfileRefResp struct {
	ID               uint                             `json:"id"`
	UserID           uint                             `json:"user_id"`
	Status           string                           `json:"status,omitempty"`
	SettlementStatus string                           `json:"settlement_status,omitempty"`
	User             *ResellerSiteConfigOwnerUserResp `json:"user,omitempty"`
}

type AdminResellerSiteConfigResp struct {
	ID           uint                              `json:"id"`
	ResellerID   uint                              `json:"reseller_id"`
	SiteName     string                            `json:"site_name"`
	Logo         string                            `json:"logo"`
	Favicon      string                            `json:"favicon"`
	Announcement models.JSON                       `json:"announcement"`
	Support      models.JSON                       `json:"support"`
	SEO          models.JSON                       `json:"seo"`
	FooterLinks  []interface{}                     `json:"footer_links"`
	NavConfig    models.JSON                       `json:"nav_config"`
	Theme        models.JSON                       `json:"theme"`
	Profile      *ResellerSiteConfigProfileRefResp `json:"profile,omitempty"`
	CreatedAt    time.Time                         `json:"created_at"`
	UpdatedAt    time.Time                         `json:"updated_at"`
}

func NewResellerProfileSummaryResp(profile *models.ResellerProfile) *ResellerProfileSummaryResp {
	if profile == nil {
		return nil
	}
	return &ResellerProfileSummaryResp{
		ID:               profile.ID,
		Status:           profile.Status,
		SettlementStatus: profile.SettlementStatus,
		CreatedAt:        profile.CreatedAt,
	}
}

func NewResellerManagementProfileResp(profile *models.ResellerProfile) *ResellerManagementProfileResp {
	if profile == nil {
		return nil
	}
	return &ResellerManagementProfileResp{
		ID:                   profile.ID,
		Status:               profile.Status,
		ApplyReason:          profile.ApplyReason,
		RejectReason:         profile.RejectReason,
		DefaultMarkupPercent: profile.DefaultMarkupPercent.String(),
		MaxMarkupPercent:     profile.MaxMarkupPercent.String(),
		SettlementStatus:     profile.SettlementStatus,
		ReviewedAt:           profile.ReviewedAt,
		CreatedAt:            profile.CreatedAt,
		UpdatedAt:            profile.UpdatedAt,
	}
}

func NewResellerDomainResp(row *models.ResellerDomain) ResellerDomainResp {
	if row == nil {
		return ResellerDomainResp{}
	}
	return ResellerDomainResp{
		ID:                 row.ID,
		Domain:             row.Domain,
		Type:               row.Type,
		VerificationToken:  row.VerificationToken,
		VerificationStatus: row.VerificationStatus,
		Status:             row.Status,
		IsPrimary:          row.IsPrimary,
		VerifiedAt:         row.VerifiedAt,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

func NewResellerDomainRespList(rows []models.ResellerDomain) []ResellerDomainResp {
	result := make([]ResellerDomainResp, 0, len(rows))
	for i := range rows {
		result = append(result, NewResellerDomainResp(&rows[i]))
	}
	return result
}

func NewResellerManagementSnapshotResp(profile *models.ResellerProfile, domains []models.ResellerDomain, canApply bool) ResellerManagementSnapshotResp {
	if profile == nil {
		return ResellerManagementSnapshotResp{Opened: false, CanApply: canApply, Domains: []ResellerDomainResp{}}
	}
	return ResellerManagementSnapshotResp{
		Opened:   true,
		CanApply: canApply,
		Profile:  NewResellerManagementProfileResp(profile),
		Domains:  NewResellerDomainRespList(domains),
	}
}

func NewResellerSiteConfigResp(row *models.ResellerSiteConfig) *ResellerSiteConfigResp {
	if row == nil {
		return nil
	}
	return &ResellerSiteConfigResp{
		ID:           row.ID,
		SiteName:     row.SiteName,
		Logo:         row.Logo,
		Favicon:      row.Favicon,
		Announcement: row.AnnouncementJSON,
		Support:      row.SupportJSON,
		SEO:          row.SEOJSON,
		FooterLinks:  resellerFooterLinksFromEnvelope(row.FooterLinksJSON),
		NavConfig:    row.NavConfigJSON,
		Theme:        row.ThemeJSON,
		UpdatedAt:    row.UpdatedAt,
	}
}

func resellerFooterLinksFromEnvelope(raw models.JSON) []interface{} {
	if raw == nil {
		return make([]interface{}, 0)
	}
	if items, ok := raw["items"].([]interface{}); ok {
		return items
	}
	if typed, ok := raw["items"].([]models.JSON); ok {
		out := make([]interface{}, 0, len(typed))
		for _, item := range typed {
			out = append(out, item)
		}
		return out
	}
	return make([]interface{}, 0)
}

func NewResellerSiteConfigSnapshotResp(profile *models.ResellerProfile, row *models.ResellerSiteConfig, canEdit bool) ResellerSiteConfigSnapshotResp {
	return ResellerSiteConfigSnapshotResp{
		Opened:  profile != nil,
		CanEdit: canEdit,
		Config:  NewResellerSiteConfigResp(row),
	}
}

func NewAdminResellerSiteConfigResp(row *models.ResellerSiteConfig) AdminResellerSiteConfigResp {
	if row == nil {
		return AdminResellerSiteConfigResp{FooterLinks: make([]interface{}, 0)}
	}
	var profile *ResellerSiteConfigProfileRefResp
	if row.Profile != nil {
		profile = &ResellerSiteConfigProfileRefResp{
			ID:               row.Profile.ID,
			UserID:           row.Profile.UserID,
			Status:           row.Profile.Status,
			SettlementStatus: row.Profile.SettlementStatus,
		}
		if row.Profile.User != nil {
			profile.User = &ResellerSiteConfigOwnerUserResp{
				ID:          row.Profile.User.ID,
				Email:       row.Profile.User.Email,
				DisplayName: row.Profile.User.DisplayName,
			}
		}
	}
	return AdminResellerSiteConfigResp{
		ID:           row.ID,
		ResellerID:   row.ResellerID,
		SiteName:     row.SiteName,
		Logo:         row.Logo,
		Favicon:      row.Favicon,
		Announcement: row.AnnouncementJSON,
		Support:      row.SupportJSON,
		SEO:          row.SEOJSON,
		FooterLinks:  resellerFooterLinksFromEnvelope(row.FooterLinksJSON),
		NavConfig:    row.NavConfigJSON,
		Theme:        row.ThemeJSON,
		Profile:      profile,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func NewAdminResellerSiteConfigRespList(rows []models.ResellerSiteConfig) []AdminResellerSiteConfigResp {
	result := make([]AdminResellerSiteConfigResp, 0, len(rows))
	for i := range rows {
		result = append(result, NewAdminResellerSiteConfigResp(&rows[i]))
	}
	return result
}

func NewResellerBalanceResp(row *models.ResellerBalanceAccount) ResellerBalanceResp {
	if row == nil {
		return ResellerBalanceResp{}
	}
	return ResellerBalanceResp{
		ID:              row.ID,
		Currency:        row.Currency,
		Status:          row.Status,
		AvailableAmount: row.AvailableAmountCache.String(),
		LockedAmount:    row.LockedAmountCache.String(),
		NegativeAmount:  row.NegativeAmountCache.String(),
		UpdatedAt:       row.UpdatedAt,
	}
}

func NewResellerBalanceRespList(rows []models.ResellerBalanceAccount) []ResellerBalanceResp {
	result := make([]ResellerBalanceResp, 0, len(rows))
	for i := range rows {
		result = append(result, NewResellerBalanceResp(&rows[i]))
	}
	return result
}

func NewResellerLedgerResp(row *models.ResellerLedgerEntry) ResellerLedgerResp {
	if row == nil {
		return ResellerLedgerResp{}
	}
	return ResellerLedgerResp{
		ID:                row.ID,
		OrderID:           row.OrderID,
		Type:              row.Type,
		Amount:            row.Amount.String(),
		Currency:          row.Currency,
		Status:            row.Status,
		AvailableAt:       row.AvailableAt,
		WithdrawRequestID: row.WithdrawRequestID,
		CreatedAt:         row.CreatedAt,
	}
}

func NewResellerLedgerRespList(rows []models.ResellerLedgerEntry) []ResellerLedgerResp {
	result := make([]ResellerLedgerResp, 0, len(rows))
	for i := range rows {
		result = append(result, NewResellerLedgerResp(&rows[i]))
	}
	return result
}

func NewResellerWithdrawResp(row *models.ResellerWithdrawRequest) ResellerWithdrawResp {
	if row == nil {
		return ResellerWithdrawResp{}
	}
	return ResellerWithdrawResp{
		ID:           row.ID,
		Amount:       row.Amount.String(),
		Currency:     row.Currency,
		Channel:      row.Channel,
		Account:      row.Account,
		Status:       row.Status,
		RejectReason: row.RejectReason,
		ProcessedAt:  row.ProcessedAt,
		CreatedAt:    row.CreatedAt,
	}
}

func NewResellerWithdrawRespList(rows []models.ResellerWithdrawRequest) []ResellerWithdrawResp {
	result := make([]ResellerWithdrawResp, 0, len(rows))
	for i := range rows {
		result = append(result, NewResellerWithdrawResp(&rows[i]))
	}
	return result
}

func NewResellerDashboardResp(opened bool, profile *models.ResellerProfile, balances []models.ResellerBalanceAccount, withdrawEnabled bool, withdrawDisabledReason string) ResellerDashboardResp {
	if !opened {
		return ResellerDashboardResp{Opened: false}
	}
	return ResellerDashboardResp{
		Opened:                 true,
		Profile:                NewResellerProfileSummaryResp(profile),
		Balances:               NewResellerBalanceRespList(balances),
		WithdrawEnabled:        withdrawEnabled,
		WithdrawDisabledReason: withdrawDisabledReason,
	}
}
