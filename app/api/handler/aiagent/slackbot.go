// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aiagent

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/easysoft/gitfox/app/api/controller/aiagent"
	"github.com/easysoft/gitfox/app/api/render"

	"github.com/slack-go/slack/slackevents"
)

func HandleSlackMessage(aiagentCtrl *aiagent.Controller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(
			body,
			// Use slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: "your_verification_token"})
			//	to verify token
			slackevents.OptionNoVerifyToken(),
		)
		if err != nil {
			http.Error(w, "Failed to parse slack event", http.StatusInternalServerError)
			return
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			if err := json.Unmarshal(body, &r); err != nil {
				http.Error(w, "Failed to parse challenge request", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte(r.Challenge))
			return
		}

		slackbotOutput, err := aiagentCtrl.HandleEvent(ctx, eventsAPIEvent)
		if err != nil {
			render.TranslatedUserError(ctx, w, err)
			return
		}

		render.JSON(w, http.StatusOK, slackbotOutput)
	}
}
