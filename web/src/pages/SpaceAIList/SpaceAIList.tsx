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
import {
  ButtonVariation,
  Container,
  FlexExpander,
  Layout,
  PageBody,
  PageHeader,
  StringSubstitute,
  TableV2 as Table,
  Text,
  useToaster
} from '@harnessio/uicore'
import { IconName } from '@harnessio/icons'
import { Color, Intent, FontVariation } from '@harnessio/design-system'
import cx from 'classnames'
import type { CellProps, Column } from 'react-table'
import { Render } from 'react-jsx-match'
import { useGet, useMutate } from 'restful-react'
import { String, useStrings } from 'framework/strings'
import { LoadingSpinner } from 'components/LoadingSpinner/LoadingSpinner'
import { SearchInputWithSpinner } from 'components/SearchInputWithSpinner/SearchInputWithSpinner'
import { NoResultCard } from 'components/NoResultCard/NoResultCard'
import { LIST_FETCHING_LIMIT, PageBrowserProps, formatDate, getErrorMessage, truncateString, voidFn } from 'utils/Utils'
import type { TypesAI } from 'services/code'
import { usePageIndex } from 'hooks/usePageIndex'
import { useQueryParams } from 'hooks/useQueryParams'
import { ResourceListingPagination } from 'components/ResourceListingPagination/ResourceListingPagination'
import { NewAIModalButton } from 'components/NewAIModalButton/NewAIModalButton'
import { useConfirmAct } from 'hooks/useConfirmAction'
import { OptionsMenuButton } from 'components/OptionsMenuButton/OptionsMenuButton'
import useUpdateAIModal from 'components/UpdateAIModal/UpdateAIModal'
import noSecretsImage from '../RepositoriesListing/no-repo.svg?url'
import css from './SpaceAIList.module.scss'
import {
  LabelListingProps,
} from 'utils/Utils'

const SpaceAIList = (props: LabelListingProps) => {
  const { activeTab, space } = props
  const { getString } = useStrings()
  const [searchTerm, setSearchTerm] = useState<string | undefined>()
  const pageBrowser = useQueryParams<PageBrowserProps>()
  const pageInit = pageBrowser.page ? parseInt(pageBrowser.page) : 1
  const [page, setPage] = usePageIndex(pageInit)

  const {
    data: aicfgs,
    error,
    loading,
    refetch,
    response
  } = useGet<TypesAI[]>({
    path: `/api/v1/spaces/${space}/+/ai`,
    queryParams: { page, limit: LIST_FETCHING_LIMIT, query: searchTerm }
  })

  const NewAIProviderButton = (
    <NewAIModalButton
      space={space}
      modalTitle={getString('ai.newProvider')}
      text={getString('ai.newProvider')}
      variation={ButtonVariation.PRIMARY}
      icon="plus"
      onSuccess={() => refetch()}></NewAIModalButton>
  )

  const { openModal: openUpdateAIModal } = useUpdateAIModal()

  const columns: Column<TypesAI>[] = useMemo(
    () => [
      {
        Header: getString('ai.provider'),
        id: 'provider',
        sort: 'true',
        width: '10%',
        Cell: ({ row }: CellProps<TypesAI>) => {
          const record = row.original
          return (
            <Text lineClamp={1}>{record?.provider}</Text>
          )
        }
      },
      {
        Header: getString('ai.endpoint'),
        id: 'endpoint',
        width: '30%',
        sort: 'true',
        Cell: ({ row }: CellProps<TypesAI>) => {
          return (
            <Layout.Horizontal style={{ alignItems: 'center' }}>
              <Text lineClamp={1}>{row.original?.endpoint}</Text>
              <Render when={row.original.is_default}>
                <Text font={{ variation: FontVariation.TINY_SEMI }} color={Color.PRIMARY_9} className={css.defaultBadge}>
                  {getString('ai.default')}
                </Text>
              </Render>
            </Layout.Horizontal>
          )
        }
      },
      {
        Header: getString('ai.model'),
        id: 'model',
        width: '20%',
        sort: 'true',
        Cell: ({ row }: CellProps<TypesAI>) => {
          return <Text lineClamp={1}>{row.original?.model}</Text>
        }
      },
      // {
      //   Header: getString('ai.default'),
      //   id: 'default',
      //   width: '10%',
      //   sort: 'false',
      //   Cell: ({ row }: CellProps<TypesAI>) => {
      //     return <Text lineClamp={1}>{row.original?.is_default ? '默认' : '-'}</Text>
      //   }
      // },
      {
        Header: getString('status'),
        id: 'status',
        width: '10%',
        sort: 'false',
        Cell: ({ row }: CellProps<TypesAI>) => {
          const iconProps = generateLastExecutionStateIcon(row.original)
          return <Text
            iconProps={{ size: 24 }}
            margin={{ right: 'medium' }}
            {...iconProps}
            tooltip={row.original.status === 'failed' ? row.original.error : undefined}
          />
        }
      },
      {
        Header: getString('created'),
        id: 'created',
        sort: 'true',
        width: '15%',
        Cell: ({ row }: CellProps<TypesAI>) => {
          return (
            <Layout.Horizontal style={{ alignItems: 'center' }}>
              <Text color={Color.BLACK} lineClamp={1} rightIconProps={{ size: 10 }} width={120}>
                {formatDate(row.original.created as number)}
              </Text>
            </Layout.Horizontal>
          )
        },
        disableSortBy: true
      },
      {
        Header: getString('lastTriggeredAt'),
        id: 'lastUsed',
        sort: 'true',
        width: '15%',
        Cell: ({ row }: CellProps<TypesAI>) => {
          return (
            <Layout.Horizontal style={{ alignItems: 'center' }}>
              <Text color={Color.BLACK} lineClamp={1} rightIconProps={{ size: 10 }} width={120}>
                {row.original.request_time && row.original.request_time !== 0
                  ? formatDate(row.original.request_time as number)
                  : '-'}
              </Text>
            </Layout.Horizontal>
          )
        },
        disableSortBy: true
      },
      {
        id: 'action',
        width: '5%',
        Cell: ({ row }: CellProps<TypesAI>) => {
          const { mutate: deleteSpaceAIConfig } = useMutate({
            verb: 'DELETE',
            path: `/api/v1/spaces/${space}/+/ai/${row.original.id}/`
          })
          const { mutate: testSpaceAIConfig } = useMutate({
            verb: 'POST',
            path: `/api/v1/spaces/${space}/+/ai/${row.original.id}/test`
          })
          const { mutate: setDefaultSpaceAIConfig } = useMutate({
            verb: 'POST',
            path: `/api/v1/spaces/${space}/+/ai/${row.original.id}/default`
          })
          const { showSuccess, showError } = useToaster()
          const confirmdeleteSpaceAIConfig = useConfirmAct()
          const confirmTestSpaceAIConfig = useConfirmAct()
          const confirmSetDefaultSpaceAIConfig = useConfirmAct()

          // TODO - add edit option
          return (
            <OptionsMenuButton
              isDark
              width="100px"
              items={[
                {
                  text: getString('edit'),
                  isDanger: true,
                  onClick: () => openUpdateAIModal({ aiToUpdate: row.original, openAIUpdate: refetch })
                },
                {
                  text: getString('test'),
                  isDanger: true,
                  onClick: () =>
                    confirmTestSpaceAIConfig({
                      title: getString('ai.testProvider'),
                      confirmText: getString('test'),
                      intent: Intent.PRIMARY,
                      message: (
                        <String
                          useRichText
                          stringID="ai.testConfirm"
                          vars={{ uid: row.original.endpoint }}
                        />
                      ),
                      action: async () => {
                        testSpaceAIConfig({})
                          .then(() => {
                            showSuccess(
                              <StringSubstitute
                                str={getString('ai.providerTestSuccess')}
                                vars={{
                                  uid: truncateString(row.original.endpoint as string, 20)
                                }}
                              />,
                              5000
                            )
                            refetch()
                          })
                          .catch(secretDeleteError => {
                            showError(getErrorMessage(secretDeleteError), 0, 'ai.providerTestFailed')
                          })
                      }
                    })
                },
                ...(row.original.is_default ? [] : [{
                  text: getString('ai.setDefaultProvider'),
                  isDanger: true,
                  onClick: () =>
                    confirmSetDefaultSpaceAIConfig({
                      title: getString('ai.setDefaultProvider'),
                      confirmText: getString('confirm'),
                      intent: Intent.PRIMARY,
                      message: (
                        <String
                          useRichText
                          stringID="ai.setDefaultConfirm"
                          vars={{ uid: row.original.endpoint }}
                        />
                      ),
                      action: async () => {
                        setDefaultSpaceAIConfig({})
                          .then(() => {
                            showSuccess(
                              <StringSubstitute
                                str={getString('ai.setDefaultProviderSuccess')}
                                vars={{
                                  uid: truncateString(row.original.endpoint as string, 20)
                                }}
                              />,
                              5000
                            )
                            refetch()
                          })
                          .catch(setDefaultProviderError => {
                            showError(getErrorMessage(setDefaultProviderError), 0, 'ai.setDefaultProviderFailed')
                          })
                      }
                    })
                }]),
                {
                  text: getString('delete'),
                  isDanger: true,
                  onClick: () =>
                    confirmdeleteSpaceAIConfig({
                      title: getString('ai.deleteProvider'),
                      confirmText: getString('delete'),
                      intent: Intent.DANGER,
                      message: (
                        <String
                          useRichText
                          stringID="ai.deleteConfirm"
                          vars={{ uid: row.original.endpoint }}
                        />
                      ),
                      action: async () => {
                        deleteSpaceAIConfig({})
                          .then(() => {
                            showSuccess(
                              <StringSubstitute
                                str={getString('ai.providerDeleted')}
                                vars={{
                                  uid: truncateString(row.original.endpoint as string, 20)
                                }}
                              />,
                              5000
                            )
                            refetch()
                          })
                          .catch(secretDeleteError => {
                            showError(getErrorMessage(secretDeleteError), 0, 'ai.providerDeleteFailed')
                          })
                      }
                    })
                }
              ]}
            />
          )
        }
      }
    ],
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [getString, refetch, searchTerm, space]
  )

  return (
    <Container className={css.main}>
      <PageBody
        className={cx({ [css.withError]: !!error })}
        error={error ? getErrorMessage(error) : null}
        retryOnError={voidFn(refetch)}
        noData={{
          when: () => aicfgs?.length === 0 && searchTerm === undefined,
          image: noSecretsImage,
          message: getString('ai.noProvidersFound'),
          button: NewAIProviderButton
        }}>
        <LoadingSpinner visible={loading && !searchTerm} />

        <Container padding="xlarge">
          <Layout.Horizontal spacing="large" className={css.layout}>
            {NewAIProviderButton}
            <FlexExpander />
            <SearchInputWithSpinner loading={loading} query={searchTerm} setQuery={setSearchTerm} />
          </Layout.Horizontal>

          <Container margin={{ top: 'medium' }}>
            {!!aicfgs?.length && (
              <Table<TypesAI>
                className={css.table}
                columns={columns}
                data={aicfgs || []}
                getRowClassName={row => cx(css.row, !row.original.description && css.noDesc)}
              />
            )}
            <NoResultCard
              showWhen={() => !!aicfgs && aicfgs?.length === 0 && !!searchTerm?.length}
              forSearch={true}
            />
          </Container>
          <ResourceListingPagination response={response} page={page} setPage={setPage} />
        </Container>
      </PageBody>
    </Container>
  )
}

const generateLastExecutionStateIcon = (
  aicfg: TypesAI
): { icon: IconName; iconProps?: { color?: Color } } => {
  let icon: IconName = 'dot'
  let color: Color | undefined = undefined

  switch (aicfg.status) {
    case 'failed':
      icon = 'danger-icon'
      break
    case 'success':
      icon = 'success-tick'
      break
    default:
      color = Color.GREY_250
  }

  return { icon, ...(color ? { iconProps: { color } } : undefined) }
}

export default SpaceAIList
