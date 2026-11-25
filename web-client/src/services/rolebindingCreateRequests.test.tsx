import {RolebindingCreateRequests} from './rolebindingCreateRequests'
import {AggregatedRoleBinding} from './role'

// adapted from sighupio/permission-manager PR #194. Additional User/Group subjects are
// now added by the backend (ADDITIONAL_SUBJECT_KINDS), so the web-client only sends the
// ServiceAccount subject.
const serviceAccountSubject = {
  kind: 'ServiceAccount',
  name: 'test-user',
  namespace: 'permission-manager'
}

describe('RolebindingCreateRequests.fromAggregatedRolebindings', () => {
  let post: jest.Mock
  let requests: RolebindingCreateRequests

  beforeEach(() => {
    post = jest.fn().mockResolvedValue({})
    requests = new RolebindingCreateRequests({post} as any)
  })

  it('creates namespaced rolebindings and the cluster rolebinding', async () => {
    const bindings: AggregatedRoleBinding[] = [
      {id: '1', template: 'template-namespaced-resources___developer', namespaces: ['dev', 'qa']}
    ]

    await requests.fromAggregatedRolebindings(bindings, 'test-user', 'read')

    expect(post).toHaveBeenCalledWith('/api/create-rolebinding', {
      roleName: 'template-namespaced-resources___developer',
      namespace: 'dev',
      roleKind: 'ClusterRole',
      subjects: [serviceAccountSubject],
      roleBindingName: 'test-user___template-namespaced-resources___developer___dev',
      generated_for_user: 'test-user'
    })
    expect(post).toHaveBeenCalledWith('/api/create-rolebinding', expect.objectContaining({namespace: 'qa'}))
    expect(post).toHaveBeenCalledWith('/api/create-cluster-rolebinding', {
      roleName: 'template-cluster-resources___read-only',
      subjects: [serviceAccountSubject],
      clusterRolebindingName: 'test-user___template-cluster-resources___read-only',
      generated_for_user: 'test-user'
    })
    expect(post).toHaveBeenCalledTimes(3)
  })

  it('creates a cluster rolebinding for ALL_NAMESPACES templates', async () => {
    const bindings: AggregatedRoleBinding[] = [
      {id: '1', template: 'template-namespaced-resources___developer', namespaces: 'ALL_NAMESPACES'}
    ]

    await requests.fromAggregatedRolebindings(bindings, 'test-user', 'none')

    expect(post).toHaveBeenCalledTimes(1)
    expect(post).toHaveBeenCalledWith('/api/create-cluster-rolebinding', {
      roleName: 'template-namespaced-resources___developer',
      subjects: [serviceAccountSubject],
      clusterRolebindingName: 'test-user___template-namespaced-resources___developerall_namespaces',
      generated_for_user: 'test-user'
    })
  })

  it('maps write cluster access to the admin template', async () => {
    await requests.fromAggregatedRolebindings([], 'test-user', 'write')

    expect(post).toHaveBeenCalledWith('/api/create-cluster-rolebinding', expect.objectContaining({
      roleName: 'template-cluster-resources___admin'
    }))
  })
})
