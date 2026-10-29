// Copyright (C) 2015 The Gravitee team (http://gravitee.io)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

func (r ListDomainsRequestObject) GetOrgId() string  { return r.OrgId }
func (r ListDomainsRequestObject) GetEnvId() string  { return r.EnvId }
func (r UpsertDomainRequestObject) GetOrgId() string { return r.OrgId }
func (r UpsertDomainRequestObject) GetEnvId() string { return r.EnvId }
func (r DeleteDomainRequestObject) GetOrgId() string { return r.OrgId }
func (r DeleteDomainRequestObject) GetEnvId() string { return r.EnvId }
func (r GetDomainRequestObject) GetOrgId() string    { return r.OrgId }
func (r GetDomainRequestObject) GetEnvId() string    { return r.EnvId }

func (r ListDataPlanesRequestObject) GetOrgId() string  { return r.OrgId }
func (r ListDataPlanesRequestObject) GetEnvId() string  { return r.EnvId }
func (r UpsertDataPlaneRequestObject) GetOrgId() string { return r.OrgId }
func (r UpsertDataPlaneRequestObject) GetEnvId() string { return r.EnvId }
func (r DeleteDataPlaneRequestObject) GetOrgId() string { return r.OrgId }
func (r DeleteDataPlaneRequestObject) GetEnvId() string { return r.EnvId }
func (r GetDataPlaneRequestObject) GetOrgId() string    { return r.OrgId }
func (r GetDataPlaneRequestObject) GetEnvId() string    { return r.EnvId }

func (r ListApplicationsRequestObject) GetOrgId() string  { return r.OrgId }
func (r ListApplicationsRequestObject) GetEnvId() string  { return r.EnvId }
func (r UpsertApplicationRequestObject) GetOrgId() string { return r.OrgId }
func (r UpsertApplicationRequestObject) GetEnvId() string { return r.EnvId }
func (r DeleteApplicationRequestObject) GetOrgId() string { return r.OrgId }
func (r DeleteApplicationRequestObject) GetEnvId() string { return r.EnvId }
func (r GetApplicationRequestObject) GetOrgId() string    { return r.OrgId }
func (r GetApplicationRequestObject) GetEnvId() string    { return r.EnvId }

func (r ListCertificatesRequestObject) GetOrgId() string  { return r.OrgId }
func (r ListCertificatesRequestObject) GetEnvId() string  { return r.EnvId }
func (r UpsertCertificateRequestObject) GetOrgId() string { return r.OrgId }
func (r UpsertCertificateRequestObject) GetEnvId() string { return r.EnvId }
func (r DeleteCertificateRequestObject) GetOrgId() string { return r.OrgId }
func (r DeleteCertificateRequestObject) GetEnvId() string { return r.EnvId }
func (r GetCertificateRequestObject) GetOrgId() string    { return r.OrgId }
func (r GetCertificateRequestObject) GetEnvId() string    { return r.EnvId }

func (r ListIdentityProvidersRequestObject) GetOrgId() string  { return r.OrgId }
func (r ListIdentityProvidersRequestObject) GetEnvId() string  { return r.EnvId }
func (r UpsertIdentityProviderRequestObject) GetOrgId() string { return r.OrgId }
func (r UpsertIdentityProviderRequestObject) GetEnvId() string { return r.EnvId }
func (r DeleteIdentityProviderRequestObject) GetOrgId() string { return r.OrgId }
func (r DeleteIdentityProviderRequestObject) GetEnvId() string { return r.EnvId }
func (r GetIdentityProviderRequestObject) GetOrgId() string    { return r.OrgId }
func (r GetIdentityProviderRequestObject) GetEnvId() string    { return r.EnvId }

func (r ListReportersRequestObject) GetOrgId() string  { return r.OrgId }
func (r ListReportersRequestObject) GetEnvId() string  { return r.EnvId }
func (r UpsertReporterRequestObject) GetOrgId() string { return r.OrgId }
func (r UpsertReporterRequestObject) GetEnvId() string { return r.EnvId }
func (r DeleteReporterRequestObject) GetOrgId() string { return r.OrgId }
func (r DeleteReporterRequestObject) GetEnvId() string { return r.EnvId }
func (r GetReporterRequestObject) GetOrgId() string    { return r.OrgId }
func (r GetReporterRequestObject) GetEnvId() string    { return r.EnvId }
