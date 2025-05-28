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
  Layout,
  Text,
  Button,
  ButtonVariation,
  Formik,
  useToaster,
  FormInput,
  FormikForm
} from '@harnessio/uicore'
import cx from 'classnames'
import { useGet, useMutate } from 'restful-react'
import { Render } from 'react-jsx-match'
import type { FormikState } from 'formik'
import { Color, FontVariation } from '@harnessio/design-system'
import type { RepoRepositoryOutput } from 'services/code'
import { useStrings } from 'framework/strings'
import { useAppContext } from 'AppContext'
import { useGetSpaceParam } from 'hooks/useGetSpaceParam'
import { VulnerabilityScanningType } from 'utils/GitUtils'
import { getErrorMessage, permissionProps } from 'utils/Utils'
import { NavigationCheck } from 'components/NavigationCheck/NavigationCheck'
import { LoadingSpinner } from 'components/LoadingSpinner/LoadingSpinner'
import css from './AISettings.module.scss'

interface AIProps {
  repoMetadata: RepoRepositoryOutput | undefined
  activeTab: string
}

interface FormData {
  aiReviewEnable: boolean
}

const AISettings = (props: AIProps) => {
  const { repoMetadata, activeTab } = props
  const { hooks, standalone, routingId } = useAppContext()
  const { getString } = useStrings()
  const { showError, showSuccess } = useToaster()
  const space = useGetSpaceParam()
  const permPushResult = hooks?.usePermissionTranslate?.(
    {
      resource: {
        resourceType: 'CODE_REPOSITORY',
        resourceIdentifier: repoMetadata?.identifier as string
      },
      permissions: ['code_repo_edit']
    },
    [space]
  )
  const { data: aiSettings, loading: aiSettingsLoading } = useGet({
    path: `/api/v1/repos/${repoMetadata?.path}/+/settings/ai`,
    queryParams: { routingId: routingId },
    lazy: !activeTab
  })
  const { mutate: updateAISettings, loading: isUpdating } = useMutate({
    verb: 'PATCH',
    path: `/api/v1/repos/${repoMetadata?.path}/+/settings/ai`,
    queryParams: { routingId: routingId }
  })

  const handleSubmit = async (
    formData: FormData,
    resetForm: (nextState?: Partial<FormikState<FormData>> | undefined) => void
  ) => {
    try {
      const payload = {
        ai_review_enabled: !!formData?.aiReviewEnable,
      }
      const response = await updateAISettings(payload)
      showSuccess(getString('aiSettings.updateSuccess'), 1500)
      resetForm({
        values: {
          aiReviewEnable: !!response?.ai_review_enabled,
        }
      })
    } catch (exception) {
      showError(getErrorMessage(exception), 1500, getString('aiSettings.failedToUpdate'))
    }
  }
  return (
    <Container className={css.main}>
      <LoadingSpinner visible={aiSettingsLoading || isUpdating} />
      {aiSettings && (
        <Formik<FormData>
          formName="aiSettings"
          initialValues={{
            aiReviewEnable: !!aiSettings?.ai_review_enabled,
          }}
          onSubmit={(formData, { resetForm }) => {
            handleSubmit(formData, resetForm)
          }}>
          {formik => {
            return (
              <FormikForm>
                <Layout.Vertical padding={{ top: 'medium' }}>
                  <Container padding="medium" margin="medium" className={css.generalContainer}>
                    <Layout.Horizontal
                      spacing={'medium'}
                      padding={{ left: 'medium' }}
                      flex={{ alignItems: 'center', justifyContent: 'flex-start' }}>
                      <FormInput.Toggle
                        {...permissionProps(permPushResult, standalone)}
                        key={'aiReviewEnable'}
                        style={{ margin: '0px' }}
                        label=""
                        name="aiReviewEnable"></FormInput.Toggle>
                      <Text className={css.title}>{getString('aiSettings.aiReview')}</Text>
                      <Text className={css.text}>{getString('aiSettings.aiReviewDesc')}</Text>
                    </Layout.Horizontal>
                  </Container>
                </Layout.Vertical>
                <Layout.Horizontal margin={'medium'} spacing={'medium'}>
                  {aiSettings?.space_ai_provider == 0 && (
                  <Button
                    variation={ButtonVariation.PRIMARY}
                    text={getString('save')}
                    onClick={() => formik.submitForm()}
                    disabled={aiSettings?.space_ai_provider == 0}
                    {...permissionProps(permPushResult, standalone)}
                  />
                  ) || (
                  <Button
                    variation={ButtonVariation.PRIMARY}
                    text={getString('save')}
                    onClick={() => formik.submitForm()}
                    disabled={formik.isSubmitting}
                    {...permissionProps(permPushResult, standalone)}
                  />
                  )}
                </Layout.Horizontal>
                <NavigationCheck when={formik.dirty} />
              </FormikForm>
            )
          }}
        </Formik>
      )}
    </Container>
  )
}

export default AISettings
