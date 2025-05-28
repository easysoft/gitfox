/*
 * Copyright 2023 Harness, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React from 'react'
import {
  Container,
  Page,
  Text,
  Layout,
} from '@harnessio/uicore'
import { Color, FontVariation } from '@harnessio/design-system'
import type { TypesAIRequest } from 'services/code'
import type { GitInfoProps } from 'utils/GitUtils'
import { useStrings } from 'framework/strings'
import { PullRequestSection } from 'utils/Utils'
import { PullRequestTabContentWrapper } from '../PullRequestTabContentWrapper'
import { MarkdownViewer } from 'components/MarkdownViewer/MarkdownViewer'
import css from './PullRequestAIReview.module.scss'

interface PullRequestAIReviewProps extends Pick<GitInfoProps, 'repoMetadata' | 'pullReqMetadata'> {
  pullReqLastAIReview?: TypesAIRequest
}

export const PullRequestAIReview: React.FC<PullRequestAIReviewProps> = ({
  repoMetadata,
  pullReqMetadata,
  pullReqLastAIReview
}) => {
  const { getString } = useStrings()
  return (
    <PullRequestTabContentWrapper section={PullRequestSection.AI_REVIEW}>
      <Container className={css.container}>
        <Page.Body>
          <Container padding="xlarge">
            <Layout.Vertical padding={{ top: 'medium' }}>
              {pullReqLastAIReview?.status === 'success' ? (
                <Container padding="large" className={css.generalContainer}>
                  <Text font={{ variation: FontVariation.H4 }} style={{ display: 'flex', alignItems: 'center' }}>
                    <span>{getString('prAIReview.reviewResult')}:</span>
                    <Text
                      color={
                        pullReqLastAIReview?.review_status === 'approved' ? Color.GREEN_700 :
                        pullReqLastAIReview?.review_status === 'rejected' ? Color.RED_700 :
                        Color.GREY_600
                      }
                      style={{ marginLeft: '8px' }}
                    >
                      {
                        pullReqLastAIReview?.review_status === 'approved' ? getString('prAIReview.approvedCode') :
                        pullReqLastAIReview?.review_status === 'rejected' ? getString('prAIReview.rejectedCode') :
                        pullReqLastAIReview?.review_status === 'ignore' ? getString('prAIReview.ignoreCode') :
                        pullReqLastAIReview?.review_status === 'unknown' ? getString('prAIReview.unknownCode') :
                        pullReqLastAIReview?.review_status
                      }
                    </Text>
                  </Text>
                  <hr className={css.dividerContainer} />
                  {pullReqLastAIReview.review_status === 'approved' ? (
                    <MarkdownViewer source={getString('prAIReview.approvedResult')} />
                  ) : pullReqLastAIReview.review_status === 'ignore' ? (
                    <MarkdownViewer source={getString('prAIReview.ignoreResult')} />
                  ) : (
                    <MarkdownViewer source={pullReqLastAIReview?.review_message || getString('prAIReview.errorResult')} />
                  )}
                </Container>
              ) : (
                <Container padding="large" className={css.generalContainer}>
                  <Text font={{ variation: FontVariation.H4 }} style={{ display: 'flex', alignItems: 'center' }}>
                    <span>{getString('prAIReview.errorResult')}</span>
                  </Text>
                  <hr className={css.dividerContainer} />
                  <MarkdownViewer source={pullReqLastAIReview?.error || getString('prAIReview.errorResultNullMessage')} />
                </Container>
              )}
            </Layout.Vertical>
          </Container>
        </Page.Body>
      </Container>
    </PullRequestTabContentWrapper>
  )
}
