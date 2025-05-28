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

import { useEffect, useMemo, useState } from 'react'
import { Color } from '@harnessio/design-system'
import { useGet } from 'restful-react'
import type { GitInfoProps } from 'utils/GitUtils'
import type { TypesAIRequest } from 'services/code'
import { ExecutionState } from 'components/ExecutionStatus/ExecutionStatus'

export function usePRAIReviews({
  repoMetadata,
  pullReqMetadata
}: Partial<Pick<GitInfoProps, 'repoMetadata' | 'pullReqMetadata'>>) {
  const { data, error, refetch } = useGet<TypesAIRequest>({
    path: `/api/v1/repos/${repoMetadata?.path}/+/pullreq/${pullReqMetadata?.number}/last-aireview`,
    queryParams: {
      debounce: 500
    }
  })
  const [color, setColor] = useState<Color>(Color.GREEN_500)
  const [background, setBackground] = useState<Color>(Color.GREEN_50)
  const [complete, setComplete] = useState(true)
  const status = useMemo(() => {
    let _status: ExecutionState | undefined
    if (data) {

      if (data.status === 'error') {
        _status = ExecutionState.ERROR
        setColor(Color.RED_900)
        setBackground(Color.RED_50)
      } else if (data.status === 'success') {
        if (data.review_status === 'approved') {
          _status = ExecutionState.SUCCESS
          setColor(Color.GREEN_800)
          setBackground(Color.GREEN_50)
        } else if (data.review_status === 'rejected') {
          _status = ExecutionState.FAILURE
          setColor(Color.RED_900)
          setBackground(Color.RED_50)
        } else {
        _status = ExecutionState.SKIPPED
        setColor(Color.GREY_600)
        setBackground(Color.GREY_100)
        }
      }
      setComplete(true)
    } else {
      setComplete(false)
    }

    return _status
  }, [data]) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    let tornDown = false
    const pollingFn = () => {
      if (repoMetadata?.path && !complete && !tornDown) {
        // TODO: fix racing condition where an ongoing refetch of the old sha overwrites the new one.
        // TEMPORARY SOLUTION: set debounce to 1 second to reduce likelyhood
        refetch({ debounce: 1 }).then(() => {
          if (!tornDown) {
            interval = window.setTimeout(pollingFn, POLLING_INTERVAL)
          }
        })
      }
    }
    let interval = window.setTimeout(pollingFn, POLLING_INTERVAL)
    return () => {
      tornDown = true
      window.clearTimeout(interval)
    }
  }, [repoMetadata?.path, complete]) // eslint-disable-line react-hooks/exhaustive-deps

  return {
    overallStatus: status,
    error,
    data,
    color,
    background,
  }
}

export type PRAIReviewsResult = ReturnType<typeof usePRAIReviews>

const POLLING_INTERVAL = 10000
