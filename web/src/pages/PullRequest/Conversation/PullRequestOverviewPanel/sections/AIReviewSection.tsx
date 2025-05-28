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
import { Container, Layout, Text } from '@harnessio/uicore'
import React, { useEffect, useState } from 'react'
import cx from 'classnames'
import { Render } from 'react-jsx-match'
import { Link } from 'react-router-dom'
import { Color, FontVariation } from '@harnessio/design-system'
import { useShowRequestError } from 'hooks/useShowRequestError'
import { useAppContext } from 'AppContext'
import type { RepoRepositoryOutput, TypesPullReq, TypesAIRequest } from 'services/code'
import { useStrings } from 'framework/strings'
import { PullRequestSection } from 'utils/Utils'
import { usePRAIReviews } from 'hooks/usePRAIReviews'
import AIStatusCircle from './AIStatusCircle'
import css from '../PullRequestOverviewPanel.module.scss'

interface AIReviewSectionProps {
  repoMetadata: RepoRepositoryOutput
  pullReqMetadata: TypesPullReq
}

const AIReviewSection = (props: AIReviewSectionProps) => {
  const { getString } = useStrings()
  const { repoMetadata, pullReqMetadata } = props
  const { routes } = useAppContext()
  const [statusMessage, setStatusMessage] = useState<{ color: string; title: string; }>({
    color: '',
    title: '',
  })
  const { data } = usePRAIReviews({ pullReqMetadata, repoMetadata })
  // useShowRequestError(error)
  const [prData, setPrData] = useState<TypesAIRequest>()

  function generateStatusSummary(checks: TypesAIRequest): { message: string; status: string } {
    // Initialize counts for each status
    if (checks.status === 'error') {
      return { message: '', status: 'error' }
    }
    if (checks.status === 'success') {
      if (checks.review_status === 'approved') {
        return { message: 'AI Review approved', status: 'success' }
      } else if (checks.review_status === 'rejected') {
        return { message: 'AI Review rejected', status: 'failed' }
      } else if (checks.review_status === 'ignore') {
        return { message: 'AI Review Ignored', status: 'skipped' }
      } else if (checks.review_status === 'unknown') {
        return { message: 'AI Review Error', status: 'killed' }
      }
    }
    return { message: 'AI Review Skipped', status: 'skipped' }
  }

  function determineStatusMessage(
    checks: TypesAIRequest
  ): { title: string; color: string } | undefined {
    let title = getString('prAIReview.oldVersion')
    let color = Color.GREY_450

    if (checks.status === 'success') {
      if (checks.review_status === 'approved') {
        return { title: getString('prAIReview.approved'), color: Color.GREEN_800 }
      } else if (checks.review_status === 'rejected') {
        return { title: getString('prAIReview.rejected'), color: Color.RED_700 }
      } else if (checks.review_status === 'ignore') {
        return { title: getString('prAIReview.ignore'), color: Color.GREY_450 }
      } else if (checks.review_status === 'unknown') {
        return { title: getString('prAIReview.unknown'), color: Color.GREY_450 }
      }
    }
    return { title, color }
  }

  useEffect(() => {
    if (data) {
      setPrData(data)
    }
  }, [data])

  useEffect(() => {
    if (prData) {
      const curStatusMessage = determineStatusMessage(prData)
      setStatusMessage(curStatusMessage || { color: '', title: '' })
    }
  }, [prData]) // eslint-disable-line react-hooks/exhaustive-deps
  return data ? (
    <Render when={data}>
      <Container
        className={cx(css.sectionContainer, css.borderContainer, { [css.mergedContainer]: pullReqMetadata.merged })}>
        <Container>
          <Layout.Horizontal flex={{ justifyContent: 'space-between' }}>
            <Layout.Horizontal flex={{ alignItems: 'center' }}>
              <Container>
                <AIStatusCircle summary={generateStatusSummary(data as TypesAIRequest)} />
              </Container>
              <Layout.Vertical padding={{ left: 'medium' }}>
                <Text
                  padding={{ bottom: 'xsmall' }}
                  className={css.sectionTitle}
                  color={(statusMessage as { color: string }).color}>
                  {statusMessage.title}
                </Text>
                <Text className={css.sectionSubheader} color={Color.GREY_450} font={{ variation: FontVariation.BODY }}>
                  {data?.review_status ? getString('prAIReview.notice') : getString('prAIReview.oldNotice')}
                </Text>
              </Layout.Vertical>
            </Layout.Horizontal>
            {data?.review_status && (
              <Link
                className={cx(css.details, css.gridItem)}
                to={
                  routes.toCODEPullRequest({
                    repoPath: repoMetadata.path as string,
                    pullRequestId: String(pullReqMetadata.number),
                    pullRequestSection: PullRequestSection.AI_REVIEW
                  })
                }>
                <Text padding={{ left: 'medium' }} color={Color.PRIMARY_7} className={css.blueText}>
                  {getString('details')}
                </Text>
              </Link>
            )}
          </Layout.Horizontal>
        </Container>
      </Container>
    </Render>
  ) : null
}

export default AIReviewSection
