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

import {
  useToaster,
  type ButtonProps,
  Button,
  Dialog,
  Layout,
  Heading,
  Container,
  Formik,
  FormikForm,
  FormInput,
  FlexExpander,
  ButtonVariation
} from '@harnessio/uicore'
import { Icon } from '@harnessio/icons'
import { Intent, FontVariation } from '@harnessio/design-system'
import React from 'react'
import { useMutate } from 'restful-react'
import * as yup from 'yup'
import { useModalHook } from 'hooks/useModalHook'
import { useStrings } from 'framework/strings'
import type { OpenapiCreateAIPluginRequest, TypesAI } from 'services/code'
import { getErrorMessage } from 'utils/Utils'
import Config from 'Config'
import css from './NewAIModalButton.module.scss'

export interface AIProviderFormData {
  endpoint: string
  model: string
  provider: string
  token: string
  showValue: boolean
}

const formInitialValues: AIProviderFormData = {
  token: '',
  provider: '',
  model: '',
  endpoint: '',
  showValue: false
}

export interface NewAIModalButtonProps extends Omit<ButtonProps, 'onClick' | 'onSubmit'> {
  space: string
  modalTitle: string
  submitButtonTitle?: string
  cancelButtonTitle?: string
  onSuccess: (ai: TypesAI) => void
}

export const NewAIModalButton: React.FC<NewAIModalButtonProps> = ({
  space,
  modalTitle,
  submitButtonTitle,
  cancelButtonTitle,
  onSuccess,
  ...props
}) => {
  const ModalComponent: React.FC = () => {
    const { getString } = useStrings()
    const { showError, showSuccess } = useToaster()

    const { mutate: createAIProvider, loading } = useMutate<TypesAI>({
      verb: 'POST',
      path: `/api/v1/spaces/${space}/ai/`
    })

    const handleSubmit = async (formData: AIProviderFormData) => {
      try {
        const payload: OpenapiCreateAIPluginRequest = {
          provider: formData.provider,
          model: formData.model,
          endpoint: formData.endpoint,
          token: formData.token,
        }
        const response = await createAIProvider(payload)
        hideModal()
        showSuccess(getString('ai.providerCreated'))
        onSuccess(response)
      } catch (exception) {
        showError(getErrorMessage(exception), 5000, getString('ai.providerUpdateFailed'))
      }
    }

    const providerDefaults = {
      openai: {
        defaultEndpoint: 'https://api.openai.com/v1',
        defaultModel: 'gpt-4o-mini',
      },
      azure: {
        defaultEndpoint: '',
        defaultModel: 'gpt-4o-mini',
      },
      deepseek: {
        defaultEndpoint: 'https://api.deepseek.com/v1',
        defaultModel: 'deepseek-chat',
      },
      gemini: {
        defaultEndpoint: 'https://api.gemini.com/v1',
        defaultModel: 'gemini-2.0-flash-exp',
      },
      anthropic: {
        defaultEndpoint: 'https://api.anthropic.com/v1',
        defaultModel: 'claude-3-5-sonnet-20241022',
      },
    } as const;

    type ProviderKey = keyof typeof providerDefaults

    return (
      <Dialog
        isOpen
        enforceFocus={false}
        onClose={hideModal}
        title={
          <Heading level={3} font={{ variation: FontVariation.H3 }}>
            {modalTitle}
          </Heading>
        }
        style={{ width: 700, maxHeight: '95vh', overflow: 'auto' }}>
        <Layout.Vertical style={{ height: '100%' }} data-testid="add-secret-modal">
          <Container>
            <Formik
              initialValues={formInitialValues}
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
                  <Container>
                    <FormInput.Select
                      name="provider"
                      label={getString('ai.provider')}
                      placeholder={getString('ai.provider')}
                      items={[
                        { label: 'OpenAI', value: 'openai' },
                        { label: 'Azure', value: 'azure' },
                        { label: 'DeepSeek', value: 'deepseek' },
                        // { label: 'Gemini', value: 'gemini' },
                        // { label: 'Anthropic', value: 'anthropic' },
                      ]}
                      usePortal
                      onChange={value => {
                        const provider = value?.value
                        if (provider && provider in providerDefaults) {
                          const providerKey = provider as ProviderKey
                          formik.setFieldValue('endpoint', providerDefaults[providerKey].defaultEndpoint)
                          formik.setFieldValue('model', providerDefaults[providerKey].defaultModel)
                        }
                      }}
                    />
                    {/* <FormInput.Text
                      name="provider"
                      label={getString('ai.provider')}
                      placeholder={getString('ai.provider')}
                      tooltipProps={{
                        dataTooltipId: 'aiProviderTextField'
                      }}
                      inputGroup={{ autoFocus: true }}
                    /> */}
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
                      inputGroup={{ autoFocus: true }}
                    />
                    <FormInput.TextArea
                      name="token"
                      label={getString('ai.token')}
                      placeholder={getString('ai.enterSecretValue')}
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
                    {/* <FormInput.Text
                      name="description"
                      label={getString('description')}
                      placeholder={getString('enterDescription')}
                      tooltipProps={{
                        dataTooltipId: 'secretDescriptionTextField'
                      }}
                      isOptional
                    /> */}
                  </Container>

                  <Layout.Horizontal
                    spacing="small"
                    padding={{ right: 'xxlarge', top: 'xxxlarge' }}
                    style={{ alignItems: 'center' }}>
                    <Button
                      type="submit"
                      text={getString('ai.createProvider')}
                      variation={ButtonVariation.PRIMARY}
                      disabled={loading}
                    />
                    <Button
                      text={cancelButtonTitle || getString('cancel')}
                      minimal
                      onClick={hideModal}
                      variation={ButtonVariation.SECONDARY}
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
  }

  const [openModal, hideModal] = useModalHook(ModalComponent, [onSuccess])

  return <Button onClick={openModal} {...props} />
}
