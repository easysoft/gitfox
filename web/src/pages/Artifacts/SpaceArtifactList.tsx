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
import { Container, Layout, PageBody, PageHeader, TableV2 as Table, Text } from '@harnessio/uicore'
import { Color } from '@harnessio/design-system'
import cx from 'classnames'
import type { CellProps, Column } from 'react-table'
import { Render } from 'react-jsx-match'
import { useGet } from 'restful-react'
import { useStrings } from 'framework/strings'
import { LoadingSpinner } from 'components/LoadingSpinner/LoadingSpinner'
import { NoResultCard } from 'components/NoResultCard/NoResultCard'
import { LIST_FETCHING_LIMIT, PageBrowserProps, getErrorMessage, voidFn, timeDistance } from 'utils/Utils'
import type { TypesArtifact } from 'services/code'
import { usePageIndex } from 'hooks/usePageIndex'
import { useQueryParams } from 'hooks/useQueryParams'
import { useGetSpaceParam } from 'hooks/useGetSpaceParam'
import { ResourceListingPagination } from 'components/ResourceListingPagination/ResourceListingPagination'
import { SpaceArtifactListHeader } from './SpaceArtifactListHeader/SpaceArtifactListHeader'
import { UserPreference, useUserPreference } from '../../hooks/useUserPreference'
import { ArtifactFilterFormatOption } from '../../utils/ArtifactUtils'
import css from './SpaceArtifactList.module.scss'

const SpaceArtifactList = () => {
  const space = useGetSpaceParam()
  const { getString } = useStrings()
  const [searchTerm, setSearchTerm] = useState<string | undefined>()
  const [filter, setFilter] = useUserPreference<string>(
    UserPreference.ARTIFACT_FORMAT_ACTIVITY_FILTER,
    ArtifactFilterFormatOption.ALL
  )
  const pageBrowser = useQueryParams<PageBrowserProps>()
  const pageInit = pageBrowser.page ? parseInt(pageBrowser.page) : 1
  const [page, setPage] = usePageIndex(pageInit)

  const {
    data: artifacts,
    error,
    loading,
    refetch,
    response
  } = useGet<TypesArtifact[]>({
    path: `/api/v1/spaces/${space}/+/artifacts`,
    queryParams: { page, limit: LIST_FETCHING_LIMIT, query: searchTerm, format: filter }
  })

  const columns: Column<TypesArtifact>[] = useMemo(
    () => [
      {
        Header: getString('artifact.format'),
        width: '100px',
        Cell: ({ row }: CellProps<TypesArtifact>) => {
          const record = row.original
          return (
            <Layout.Horizontal spacing="small" style={{ flexGrow: 1 }}>
              {record.format && <Text lineClamp={1}>{record.format}</Text>}
            </Layout.Horizontal>
          )
        }
      },
      {
        Header: getString('artifact.display_name'),
        width: '640px',
        Cell: ({ row }: CellProps<TypesArtifact>) => {
          return (
            <Layout.Horizontal style={{ alignItems: 'center' }}>
              <Text color={Color.BLACK} lineClamp={1} rightIconProps={{ size: 10 }} width={320}>
                {row.original.display_name}
              </Text>
            </Layout.Horizontal>
          )
        },
        disableSortBy: true
      },
      {
        Header: getString('artifact.version'),
        width: '140px',
        Cell: ({ row }: CellProps<TypesArtifact>) => {
          return (
            <Layout.Horizontal style={{ alignItems: 'center' }}>
              <Text color={Color.BLACK} lineClamp={1} rightIconProps={{ size: 10 }} width={140}>
                {row.original.version}
              </Text>
            </Layout.Horizontal>
          )
        },
        disableSortBy: true
      },
      {
        Header: getString('artifact.update_time'),
        width: '180px',
        Cell: ({ row }: CellProps<TypesArtifact>) => {
          return (
            <Layout.Horizontal style={{ alignItems: 'center' }}>
              <Text color={Color.BLACK} lineClamp={1} rightIconProps={{ size: 10 }} width={120}>
                {timeDistance(Date.now(), row.original.update_time, true)}
              </Text>
            </Layout.Horizontal>
          )
        },
        disableSortBy: true
      }
    ],
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [getString, refetch, searchTerm, space]
  )

  return (
    <Container className={css.main}>
      <PageHeader title={getString('pageTitle.artifacts')} className="PageHeader--container PageHeader--standard" />

      <PageBody error={error ? getErrorMessage(error) : null} retryOnError={voidFn(refetch)}>
        <LoadingSpinner visible={loading && !searchTerm} />

        <Render when={true}>
          <Layout.Vertical>
            <SpaceArtifactListHeader
              loading={loading}
              activeFormatFilterOption={filter}
              onArtifactFormatFilterChanged={_filter => {
                setFilter(_filter)
                setPage(1)
              }}
              onSearchTermChanged={value => {
                setSearchTerm(value)
                setPage(1)
              }}
            />
            <Container padding="xlarge">
              <Container margin={{ top: 'medium' }}>
                {!!artifacts?.length && (
                  <Table<TypesArtifact>
                    className={css.table}
                    columns={columns}
                    data={artifacts || []}
                    getRowClassName={row => cx(css.row, !row.original.display_name)}
                  />
                )}
                <NoResultCard
                  showWhen={() => !!artifacts && artifacts?.length === 0 && !!searchTerm?.length}
                  forSearch={true}
                />
              </Container>
              <ResourceListingPagination response={response} page={page} setPage={setPage} />
            </Container>
          </Layout.Vertical>
        </Render>
      </PageBody>
    </Container>
  )
}

export default SpaceArtifactList
