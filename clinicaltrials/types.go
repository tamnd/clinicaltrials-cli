package clinicaltrials

import "strings"

// Study is the record emitted for search and study operations.
type Study struct {
	NCTID          string   `kit:"id" json:"nct_id"`
	BriefTitle     string   `json:"brief_title"`
	OfficialTitle  string   `json:"official_title"`
	OverallStatus  string   `json:"overall_status"`
	Phase          string   `json:"phase"`
	StudyType      string   `json:"study_type"`
	Enrollment     int      `json:"enrollment"`
	StartDate      string   `json:"start_date"`
	CompletionDate string   `json:"completion_date"`
	LeadSponsor    string   `json:"lead_sponsor"`
	Conditions     []string `json:"conditions"`
	BriefSummary   string   `json:"brief_summary"`
}

// ─── wire types ──────────────────────────────────────────────────────────────

type wireListResponse struct {
	Studies       []wireStudy `json:"studies"`
	NextPageToken string      `json:"nextPageToken"`
}

type wireStudy struct {
	ProtocolSection struct {
		IdentificationModule struct {
			NCTID         string `json:"nctId"`
			BriefTitle    string `json:"briefTitle"`
			OfficialTitle string `json:"officialTitle"`
		} `json:"identificationModule"`
		StatusModule struct {
			OverallStatus        string `json:"overallStatus"`
			StartDateStruct      struct{ Date string `json:"date"` } `json:"startDateStruct"`
			CompletionDateStruct struct{ Date string `json:"date"` } `json:"completionDateStruct"`
		} `json:"statusModule"`
		SponsorCollaboratorsModule struct {
			LeadSponsor struct{ Name string `json:"name"` } `json:"leadSponsor"`
		} `json:"sponsorCollaboratorsModule"`
		DescriptionModule struct {
			BriefSummary string `json:"briefSummary"`
		} `json:"descriptionModule"`
		ConditionsModule struct {
			Conditions []string `json:"conditions"`
		} `json:"conditionsModule"`
		DesignModule struct {
			StudyType      string   `json:"studyType"`
			Phases         []string `json:"phases"`
			EnrollmentInfo struct {
				Count int `json:"count"`
			} `json:"enrollmentInfo"`
		} `json:"designModule"`
	} `json:"protocolSection"`
}

// toStudy converts a wireStudy to a Study.
func toStudy(w wireStudy) Study {
	ps := w.ProtocolSection
	phase := ""
	if len(ps.DesignModule.Phases) > 0 {
		phase = ps.DesignModule.Phases[0]
	}
	summary := strings.ReplaceAll(ps.DescriptionModule.BriefSummary, "\n", " ")
	summary = strings.TrimSpace(summary)
	return Study{
		NCTID:          ps.IdentificationModule.NCTID,
		BriefTitle:     ps.IdentificationModule.BriefTitle,
		OfficialTitle:  ps.IdentificationModule.OfficialTitle,
		OverallStatus:  ps.StatusModule.OverallStatus,
		Phase:          phase,
		StudyType:      ps.DesignModule.StudyType,
		Enrollment:     ps.DesignModule.EnrollmentInfo.Count,
		StartDate:      ps.StatusModule.StartDateStruct.Date,
		CompletionDate: ps.StatusModule.CompletionDateStruct.Date,
		LeadSponsor:    ps.SponsorCollaboratorsModule.LeadSponsor.Name,
		Conditions:     ps.ConditionsModule.Conditions,
		BriefSummary:   summary,
	}
}
