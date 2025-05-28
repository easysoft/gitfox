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

import React, { useMemo, useState } from 'react'
import { Container, Layout, FlexExpander, DropDown } from '@harnessio/uicore'
import { ArtifactFilterFormatOption } from 'utils/ArtifactUtils'
import { UserPreference, useUserPreference } from 'hooks/useUserPreference'
import { SearchInputWithSpinner } from 'components/SearchInputWithSpinner/SearchInputWithSpinner'
import css from './SpaceArtifactListHeader.module.scss'

interface SpaceArtifactListHeaderProps {
  loading?: boolean
  activeFormatFilterOption?: string
  onArtifactFormatFilterChanged: (filter: string) => void
  onSearchTermChanged: (searchTerm: string) => void
}

export function SpaceArtifactListHeader({
  loading,
  onArtifactFormatFilterChanged,
  onSearchTermChanged,
  activeFormatFilterOption = ArtifactFilterFormatOption.ALL
}: SpaceArtifactListHeaderProps) {
  const [filterOption, setFilterOption] = useUserPreference(
    UserPreference.ARTIFACT_FORMAT_ACTIVITY_FILTER,
    activeFormatFilterOption
  )

  const [searchTerm, setSearchTerm] = useState('')

  const items = useMemo(
    () => [
      { label: 'raw', value: ArtifactFilterFormatOption.RAW },
      { label: 'helm', value: ArtifactFilterFormatOption.HELM },
      { label: 'container', value: ArtifactFilterFormatOption.CONTAINER },
      { label: 'all', value: ArtifactFilterFormatOption.ALL }
    ],
    []
  )

  return (
    <Container className={css.main} padding="xlarge">
      <Layout.Horizontal spacing="medium">
        <SearchInputWithSpinner
          loading={loading}
          spinnerPosition="right"
          query={searchTerm}
          setQuery={value => {
            setSearchTerm(value)
            onSearchTermChanged(value)
          }}
        />
        <FlexExpander />
        <DropDown
          value={filterOption}
          items={items}
          onChange={({ value }) => {
            setFilterOption(value as string)
            onArtifactFormatFilterChanged(value as string)
          }}
          popoverClassName={css.branchDropdown}
        />
      </Layout.Horizontal>
    </Container>
  )
}
