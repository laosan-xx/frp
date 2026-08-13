import { buildQueryString, http } from './http'
import type { V2Page } from './http'
import type { ClientInfoData, ClientListV2Params } from '../types/client'

export interface SystemPruneResponse {
  type: string
  cleared: number
  total: number
}

export const getClients = () => {
  return http.get<ClientInfoData[]>('../api/clients')
}

export const getClientsV2 = (params: ClientListV2Params = {}) => {
  return http.getV2<V2Page<ClientInfoData>>(
    `../api/v2/clients${buildQueryString({
      page: params.page,
      pageSize: params.pageSize,
      status:
        params.status && params.status !== 'all' ? params.status : undefined,
      q: params.q || undefined,
      clientID: params.clientID || undefined,
      runID: params.runID || undefined,
    })}`,
  )
}

export const getClient = (key: string) => {
  return http.get<ClientInfoData>(`../api/clients/${key}`)
}

export const getClientV2 = (key: string) => {
  return http.getV2<ClientInfoData>(`../api/v2/clients/${encodeURIComponent(key)}`)
}

export const deleteClientV2 = (key: string) => {
  return http.deleteV2<{ status: string }>(
    `../api/v2/clients/${encodeURIComponent(key)}`,
  )
}

// Resolve a client by clientID (preferred) or runID, without the
// composite "{user}.{clientID}" key. Supports the clientID-based routing.
export const getClientByID = (id: string) => {
  return http.getV2<ClientInfoData>(`../api/v2/client/${encodeURIComponent(id)}`)
}

// Resolve a client by runID (legacy devices without a clientID).
export const getClientByRunID = (runID: string) => {
  return http.getV2<ClientInfoData>(
    `../api/v2/client/run/${encodeURIComponent(runID)}`,
  )
}

// Delete an offline client identified by clientID or runID.
export const deleteClientByID = (id: string) => {
  return http.deleteV2<{ status: string }>(
    `../api/v2/client/${encodeURIComponent(id)}`,
  )
}

export const clearOfflineClients = () => {
  return http.postV2<SystemPruneResponse>(
    '../api/v2/system/prune?type=offline_clients',
  )
}

export interface ClientCommandReq {
  command: string
  payload: string
}

export interface ClientCommandResp {
  command: string
  result: string
  output: string
}

export const sendClientCommand = (key: string, req: ClientCommandReq) => {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 35000)
  return http
    .postV2<ClientCommandResp>(
      `../api/v2/clients/${encodeURIComponent(key)}/command`,
      req,
      { signal: controller.signal },
    )
    .finally(() => clearTimeout(timeoutId))
}

// Send a command to a client identified by clientID or runID.
export const sendClientCommandByID = (id: string, req: ClientCommandReq) => {
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 35000)
  return http
    .postV2<ClientCommandResp>(
      `../api/v2/client/${encodeURIComponent(id)}/command`,
      req,
      { signal: controller.signal },
    )
    .finally(() => clearTimeout(timeoutId))
}

// Firmware releases (fetched server-side from GitHub, bypasses shared proxy rate limiting)
export interface FirmwareAsset {
  name: string
  size: number
  url: string
}

export interface FirmwareBranch {
  branch: string
  config: string
  date: string
  assets: FirmwareAsset[]
}

export interface FirmwareReleasesResp {
  branches: FirmwareBranch[]
}

export const fetchFirmwareReleases = (repoApi: string, boardModel: string) => {
  return http.getV2<FirmwareReleasesResp>(
    `../api/v2/firmware/releases${buildQueryString({ repoApi, boardModel })}`,
  )
}
