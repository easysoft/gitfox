export enum ArtifactFormat {
  RAW = 'raw',
  HELM = 'helm',
  CONTAINER = 'container'
}

export const ArtifactFilterFormatOption = {
  ...ArtifactFormat,
  ALL: ''
}
