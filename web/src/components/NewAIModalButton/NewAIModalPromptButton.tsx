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
import type { OpenapiCreateAIPromptRequest, TypesAIPrompt } from 'services/code'
import { getErrorMessage } from 'utils/Utils'
import Config from 'Config'
import css from './NewAIModalButton.module.scss'

export interface AIPromptFormData {
  prompt: string
  description: string
}

const formInitialValues: AIPromptFormData = {
  prompt: '',
  description: ''
}

export interface NewAIModalButtonProps extends Omit<ButtonProps, 'onClick' | 'onSubmit'> {
  space: string
  modalTitle: string
  submitButtonTitle?: string
  cancelButtonTitle?: string
  onSuccess: (ai: TypesAIPrompt) => void
}

export const NewAIModalPromptButton: React.FC<NewAIModalButtonProps> = ({
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

    const { mutate: createAIPrompt, loading } = useMutate<TypesAIPrompt>({
      verb: 'POST',
      path: `/api/v1/spaces/${space}/aiprompt/`
    })

    const handleSubmit = async (formData: AIPromptFormData) => {
      try {
        const payload: OpenapiCreateAIPromptRequest = {
          prompt: formData.prompt,
          description: formData.description,
        }
        const response = await createAIPrompt(payload)
        hideModal()
        showSuccess(getString('ai.providerCreated'))
        onSuccess(response)
      } catch (exception) {
        showError(getErrorMessage(exception), 5000, getString('ai.providerUpdateFailed'))
      }
    }

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
        <Layout.Vertical style={{ height: '100%' }} data-testid="add-prompt-modal">
          <Container>
            <Formik
              initialValues={formInitialValues}
              formName="addAIPrompt"
              enableReinitialize={true}
              validationSchema={yup.object().shape({
                prompt: yup
                  .string()
                  .trim()
                  .required(getString('validation.required', { name: getString('ai.prompt') })) // GITFOX!
              })}
              validateOnChange
              validateOnBlur
              onSubmit={handleSubmit}>
              {formik => (
                <FormikForm>
                  <Container>
                    {/* <FormInput.Select
                      name="provider"
                      label={getString('ai.provider')}
                      placeholder={getString('ai.provider')}
                      items={[
                        { label: 'OpenAI', value: 'openai' },
                        { label: 'Azure', value: 'azure' },
                        { label: 'DeepSeek', value: 'deepseek' },
                      ]}
                      usePortal
                    /> */}
                    {/* <FormInput.Text
                      name="provider"
                      label={getString('ai.provider')}
                      placeholder={getString('ai.provider')}
                      tooltipProps={{
                        dataTooltipId: 'aiProviderTextField'
                      }}
                      inputGroup={{ autoFocus: true }}
                    /> */}
                    <FormInput.TextArea
                      name="prompt"
                      label={getString('ai.prompt')}
                      placeholder={getString('ai.prompt')}
                      tooltipProps={{
                        dataTooltipId: 'aiPromptTextField'
                      }}
                      maxLength={Config.SECRET_LIMIT_IN_BYTES}
                      autoComplete="off"
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
