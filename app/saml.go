// Copyright (c) 2017-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/mattermost/mattermost-server/model"
)

const (
	SamlPublicCertificateName = "saml-public.crt"
	SamlPrivateKeyName        = "saml-private.key"
	SamlIdpCertificateName    = "saml-idp.crt"
)

func (a *App) GetSamlMetadata() (string, *model.AppError) {
	if a.Saml == nil {
		err := model.NewAppError("GetSamlMetadata", "api.admin.saml.not_available.app_error", nil, "", http.StatusNotImplemented)
		return "", err
	}

	result, err := a.Saml.GetMetadata()
	if err != nil {
		return "", model.NewAppError("GetSamlMetadata", "api.admin.saml.metadata.app_error", nil, "err="+err.Message, err.StatusCode)
	}
	return result, nil
}

func (a *App) writeSamlFile(name string, fileData *multipart.FileHeader) *model.AppError {
	file, err := fileData.Open()
	if err != nil {
		return model.NewAppError("AddSamlCertificate", "api.admin.add_certificate.open.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return model.NewAppError("AddSamlCertificate", "api.admin.add_certificate.saving.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	err = a.Srv.configStore.SetFile(name, data)
	if err != nil {
		return model.NewAppError("AddSamlCertificate", "api.admin.add_certificate.saving.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) AddSamlPublicCertificate(fileData *multipart.FileHeader) *model.AppError {
	if err := a.writeSamlFile(SamlPublicCertificateName, fileData); err != nil {
		return err
	}

	cfg := a.Config().Clone()
	*cfg.SamlSettings.PublicCertificateFile = SamlPublicCertificateName

	if err := cfg.IsValid(); err != nil {
		return err
	}

	a.UpdateConfig(func(dest *model.Config) { *dest = *cfg })
	a.PersistConfig()

	return nil
}

func (a *App) AddSamlPrivateCertificate(fileData *multipart.FileHeader) *model.AppError {
	if err := a.writeSamlFile(SamlPrivateKeyName, fileData); err != nil {
		return err
	}

	a.UpdateConfig(func(cfg *model.Config) {
		*cfg.SamlSettings.PrivateKeyFile = SamlPrivateKeyName
	})
	a.PersistConfig()

	return nil
}

func (a *App) AddSamlIdpCertificate(fileData *multipart.FileHeader) *model.AppError {
	if err := a.writeSamlFile(SamlIdpCertificateName, fileData); err != nil {
		return err
	}

	a.UpdateConfig(func(cfg *model.Config) {
		*cfg.SamlSettings.IdpCertificateFile = SamlIdpCertificateName
	})
	a.PersistConfig()

	return nil
}

func (a *App) RemoveSamlPublicCertificate() *model.AppError {
	if err := a.Srv.configStore.RemoveFile(SamlPublicCertificateName); err != nil {
		return model.NewAppError("RemoveSamlFile", "api.admin.remove_certificate.delete.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	a.UpdateConfig(func(cfg *model.Config) {
		*cfg.SamlSettings.PublicCertificateFile = ""
		*cfg.SamlSettings.Encrypt = false
	})
	a.PersistConfig()

	return nil
}

func (a *App) RemoveSamlPrivateCertificate() *model.AppError {
	if err := a.Srv.configStore.RemoveFile(SamlPrivateKeyName); err != nil {
		return model.NewAppError("RemoveSamlFile", "api.admin.remove_certificate.delete.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	a.UpdateConfig(func(cfg *model.Config) {
		*cfg.SamlSettings.PrivateKeyFile = ""
		*cfg.SamlSettings.Encrypt = false
	})
	a.PersistConfig()

	return nil
}

func (a *App) RemoveSamlIdpCertificate() *model.AppError {
	if err := a.Srv.configStore.RemoveFile(SamlIdpCertificateName); err != nil {
		return model.NewAppError("RemoveSamlFile", "api.admin.remove_certificate.delete.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	a.UpdateConfig(func(cfg *model.Config) {
		*cfg.SamlSettings.IdpCertificateFile = ""
		*cfg.SamlSettings.Enable = false
	})
	a.PersistConfig()

	return nil
}

func (a *App) GetSamlCertificateStatus() *model.SamlCertificateStatus {
	status := &model.SamlCertificateStatus{}

	status.IdpCertificateFile, _ = a.Srv.configStore.HasFile(SamlIdpCertificateName)
	status.PrivateKeyFile, _ = a.Srv.configStore.HasFile(SamlPrivateKeyName)
	status.PublicCertificateFile, _ = a.Srv.configStore.HasFile(SamlPublicCertificateName)

	return status
}
