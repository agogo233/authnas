export * from './users'
export * from './groups'
export * from './clients'
export * from './invitations'
export * from './proxyauth'

import { adminUsersApi } from './users'
import { adminGroupsApi } from './groups'
import { adminClientsApi } from './clients'
import { adminInvitationsApi } from './invitations'
import { adminProxyAuthApi } from './proxyauth'

export const adminApi = {
  users: adminUsersApi,
  groups: adminGroupsApi,
  clients: adminClientsApi,
  invitations: adminInvitationsApi,
  proxyauth: adminProxyAuthApi,
}
