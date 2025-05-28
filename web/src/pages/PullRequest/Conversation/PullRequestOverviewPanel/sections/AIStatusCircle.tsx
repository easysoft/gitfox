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
import { Container } from '@harnessio/uicore'
import { useStrings } from 'framework/strings'
import { ExecutionState, ExecutionStatus } from 'components/ExecutionStatus/ExecutionStatus'
import Success from '../../../../../icons/code-success.svg?url'
import css from '../PullRequestOverviewPanel.module.scss'

// Define the AIStatusCircle component
const AIStatusCircle = ({
  summary
}: {
  summary: {
    message: string
    status: string
  }
}) => {
  const { getString } = useStrings()
  const status = summary.status
  return (
    <>
      {status === 'success' ? (
        <img alt={getString('success')} width={27} height={27} src={Success} />
      ) : (
        <Container className={css.statusCircleContainer}>
          <ExecutionStatus
            className={css.iconStatus}
            status={status as ExecutionState}
            iconOnly
            noBackground
            iconSize={status === 'error' ? 27 : 26}
            isCi
            inPr
          />
        </Container>
      )}
    </>
  )
}

export default AIStatusCircle
