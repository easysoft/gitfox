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

import React, { useRef, useState } from 'react'
import * as yup from 'yup'
import { useMutate } from 'restful-react'
import { FontVariation, Intent } from '@harnessio/design-system'
import {
  Button,
  Dialog,
  Layout,
  Heading,
  Container,
  Formik,
  FormikForm,
  FormInput,
  FlexExpander,
  useToaster,
  StringSubstitute,
  ButtonVariation
} from '@harnessio/uicore'
import { Icon } from '@harnessio/icons'
import { useStrings } from 'framework/strings'
import { useModalHook } from 'hooks/useModalHook'
import { useGetSpaceParam } from 'hooks/useGetSpaceParam'
import type { OpenapiUpdateAIProviderRequest, TypesAI } from 'services/code'
import type { AIProviderFormData } from 'components/NewAIModalButton/NewAIModalButton'
import { getErrorMessage, truncateString } from 'utils/Utils'
import Config from 'Config'
import css from './UpdateAIModal.module.scss'

const useUpdateAIModal = () => {
  const { getString } = useStrings()
  const space = useGetSpaceParam()
  const { showError, showSuccess } = useToaster()
  const [ai, setSecret] = useState<TypesAI>()
  const postUpdate = useRef<() => Promise<void>>()

  const { mutate: updateSecret, loading } = useMutate<TypesAI>({
    verb: 'PATCH',
    path: `/api/v1/spaces/${space}/ai/${ai?.id}/`
  })

  const handleSubmit = async (formData: AIProviderFormData) => {
    try {
      const payload: OpenapiUpdateAIProviderRequest = {
        endpoint: formData.endpoint,
        model: formData.model,
        token: formData.token,
      }
      await updateSecret(payload)
      hideModal()
      showSuccess(
        <StringSubstitute
          str={getString('ai.providerUpdated')}
          vars={{
            uid: truncateString(formData.endpoint, 20)
          }}
        />
      )
      postUpdate.current?.()
    } catch (exception) {
      showError(getErrorMessage(exception), 0, getString('ai.providerUpdateFailed'))
    }
  }

  const [openModal, hideModal] = useModalHook(() => {
    const onClose = () => {
      hideModal()
    }

    return (
      <Dialog
        isOpen
        enforceFocus={false}
        onClose={hideModal}
        title={
          <Heading level={3} font={{ variation: FontVariation.H3 }}>
            {getString('ai.providerUpdated')}
          </Heading>
        }
        style={{ width: 700, maxHeight: '95vh', overflow: 'auto' }}>
        <Layout.Vertical style={{ height: '100%' }} data-testid="add-secret-modal">
          <Container>
            <Formik
              initialValues={{
                provider: ai?.provider || '',
                endpoint: ai?.endpoint || '',
                model: ai?.model || '',
                token: '',
                showValue: false
              }}
              formName="addAIProvider"
              enableReinitialize={true}
              validationSchema={yup.object().shape({
                endpoint: yup
                  .string()
                  .trim()
                  .required(getString('validation.required', { name: getString('ai.endpoint') })) // GITFOX!
                  .matches(/^https?:\/\/.*$/, getString('validation.endpointLogic')),
                model: yup
                  .string()
                  .trim()
                  .required(getString('validation.required', { name: getString('ai.model') })) // GITFOX!
                  .min(3, getString('validation.nameTooShort'))
                  .max(30, getString('validation.nameTooLong'))
                  .matches(/^[a-zA-Z][a-zA-Z0-9]*([-][a-zA-Z0-9]+)*$/, getString('validation.modelLogic')),
                token: yup
                  .string()
                  .trim()
                  .required(getString('validation.required', { name: getString('ai.token') })) // GITFOX!
              })}
              validateOnChange
              validateOnBlur
              onSubmit={handleSubmit}>
              {formik => (
                <FormikForm>
                  <FormInput.Text
                    name="endpoint"
                    label={getString('ai.endpoint')}
                    placeholder={getString('ai.endpoint')}
                    tooltipProps={{
                      dataTooltipId: 'aiEndpointTextField'
                    }}
                    inputGroup={{ autoFocus: true }}
                  />
                  <FormInput.Text
                    name="model"
                    label={getString('ai.model')}
                    placeholder={getString('ai.model')}
                    tooltipProps={{
                      dataTooltipId: 'aiModelTextField'
                    }}
                   // isOptional
                  />
                  <FormInput.TextArea
                    name="token"
                    label={getString('ai.token')}
                    placeholder={getString('secrets.enterSecretValue')}
                    tooltipProps={{
                      dataTooltipId: 'aiTokenTextField'
                    }}
                    maxLength={Config.SECRET_LIMIT_IN_BYTES}
                    autoComplete="off"
                    className={formik.values.showValue ? css.showValue : css.hideValue}
                  />
                  <FormInput.CheckBox
                    name="showValue"
                    label={getString('ai.showValue')}
                    tooltipProps={{
                      dataTooltipId: 'aiTokenTextField'
                    }}
                    style={{ display: 'flex' }}
                  />
                  <Layout.Horizontal
                    spacing="small"
                    padding={{ right: 'xxlarge', top: 'xxxlarge' }}
                    style={{ alignItems: 'center' }}>
                    <Button
                      type="submit"
                      text={getString('ai.providerUpdated')}
                      variation={ButtonVariation.PRIMARY}
                      disabled={loading}
                    />
                    <Button
                      text={getString('cancel')}
                      minimal
                      variation={ButtonVariation.SECONDARY}
                      onClick={onClose}
                    />
                    <FlexExpander />
                    {loading && <Icon intent={Intent.PRIMARY} name="steps-spinner" size={16} />}
                  </Layout.Horizontal>
                </FormikForm>
              )}
            </Formik>
          </Container>
        </Layout.Vertical>
      </Dialog>
    )
  }, [ai])

  return {
    openModal: ({
      aiToUpdate,
      openAIUpdate
    }: {
      aiToUpdate: TypesAI
      openAIUpdate: () => Promise<void>
    }) => {
      setSecret(aiToUpdate)
      postUpdate.current = openAIUpdate
      openModal()
    }
  }
}

export default useUpdateAIModal
