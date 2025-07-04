/*
 * Copyright 2024 Harness, Inc.
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
import { Container, Layout } from '@harnessio/uicore'
import { Link } from 'react-router-dom'
import { useAppContext } from 'AppContext'
import harness from './gitfox.svg?url'
import css from './HarnessLogo.module.scss'

export const HarnessLogo: React.FC = () => {
  const { routes } = useAppContext()

  return (
    <Container className={css.main}>
      <Link to={routes.toCODEHome()}>
        <Layout.Horizontal spacing="small" className={css.layout} padding={{ left: 'small' }}>
          <img src={harness} width={100} height={50} />
        </Layout.Horizontal>
      </Link>
    </Container>
  )
}
