// Copyright 2026 Sloth. All rights reserved.
// License can be found in the LICENSE file.

package api

import (
	"net/http"
	"strings"

	"github.com/clivern/sloth/db"
	"github.com/clivern/sloth/locale"
	"github.com/clivern/sloth/module"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// ListCatalogSkillsAction lists the skills the calling agent may call.
func (a *API) ListCatalogSkillsAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	caller, _ := module.AgentFrom(r.Context())

	limit, offset := a.ParsePagination(r)
	query := r.URL.Query()

	result, err := a.Catalog.ListSkills(db.Id(wid), caller, &module.ListCatalogSkillsRequest{
		Tags:    lo.Compact(strings.Split(query.Get("tags"), ",")),
		Passing: query.Get("passing") == "true",
		Limit:   limit,
		Offset:  offset,
	})

	if err != nil {
		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", caller.Name).
			Msg("Failed to list catalog skills")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_catalog_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"skills": result.Skills,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}

// ListCatalogSkillAgentsAction lists the agents the calling agent may call for a skill.
func (a *API) ListCatalogSkillAgentsAction(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "workspaceId")
	skill := chi.URLParam(r, "skill")
	caller, _ := module.AgentFrom(r.Context())

	limit, offset := a.ParsePagination(r)

	result, err := a.Catalog.ListSkillAgents(db.Id(wid), caller, skill, &module.ListCatalogSkillAgentsRequest{
		Passing: r.URL.Query().Get("passing") == "true",
		Limit:   limit,
		Offset:  offset,
	})

	if err != nil {
		if a.WriteError(w, err, []ErrorMapping{
			{module.ErrSkillNotFound, http.StatusNotFound, locale.TR(r, "skill_not_found")},
		}) {
			return
		}

		log.Error().
			Err(err).
			Str("workspaceId", wid).
			Str("agent", caller.Name).
			Str("skill", skill).
			Msg("Failed to list catalog skill agents")
		a.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"errorMessage": locale.TR(r, "failed_catalog_request"),
		})

		return
	}

	a.WriteJSON(w, http.StatusOK, map[string]any{
		"skill":  result.Skill,
		"agents": result.Agents,
		"_meta": map[string]any{
			"limit":  limit,
			"offset": offset,
			"total":  result.Total,
		},
	})
}
