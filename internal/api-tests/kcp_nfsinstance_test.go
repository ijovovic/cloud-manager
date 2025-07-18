package api_tests

import (
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Feature: KCP NfsInstance", func() {

	// OpenStack ============================================

	It("Scenario: KCP NfsInstance SAP without IpRange can be created", func() {
		name := uuid.NewString()
		var err error
		obj := &cloudcontrolv1beta1.NfsInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: infra.KCP().Namespace(),
			},
			Spec: cloudcontrolv1beta1.NfsInstanceSpec{
				Scope: cloudcontrolv1beta1.ScopeRef{Name: "s"},
				Instance: cloudcontrolv1beta1.NfsInstanceInfo{
					OpenStack: &cloudcontrolv1beta1.NfsInstanceOpenStack{
						SizeGb: 10,
					},
				},
			},
		}

		err = infra.KCP().Client().Create(infra.Ctx(), obj)
		Expect(err).NotTo(HaveOccurred())

		_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
	})

	It("Scenario: KCP NfsInstance SAP with IpRange can not be created", func() {
		name := uuid.NewString()
		var err error
		obj := &cloudcontrolv1beta1.NfsInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: infra.KCP().Namespace(),
			},
			Spec: cloudcontrolv1beta1.NfsInstanceSpec{
				Scope: cloudcontrolv1beta1.ScopeRef{Name: "s"},
				IpRange: cloudcontrolv1beta1.IpRangeRef{
					Name: "foo",
				},
				Instance: cloudcontrolv1beta1.NfsInstanceInfo{
					OpenStack: &cloudcontrolv1beta1.NfsInstanceOpenStack{
						SizeGb: 10,
					},
				},
			},
		}

		err = infra.KCP().Client().Create(infra.Ctx(), obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("IpRange can not be specified for openstack, and is mandatory for gcp and aws"))

		_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
	})

	// AWS ============================================

	It("Scenario: KCP NfsInstance AWS without IpRange can not be created", func() {
		name := uuid.NewString()
		var err error
		obj := &cloudcontrolv1beta1.NfsInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: infra.KCP().Namespace(),
			},
			Spec: cloudcontrolv1beta1.NfsInstanceSpec{
				Scope: cloudcontrolv1beta1.ScopeRef{Name: "s"},
				Instance: cloudcontrolv1beta1.NfsInstanceInfo{
					Aws: &cloudcontrolv1beta1.NfsInstanceAws{},
				},
			},
		}

		err = infra.KCP().Client().Create(infra.Ctx(), obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("IpRange can not be specified for openstack, and is mandatory for gcp and aws"))

		_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
	})

	It("Scenario: KCP NfsInstance AWS with IpRange can be created", func() {
		name := uuid.NewString()
		var err error
		obj := &cloudcontrolv1beta1.NfsInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: infra.KCP().Namespace(),
			},
			Spec: cloudcontrolv1beta1.NfsInstanceSpec{
				Scope: cloudcontrolv1beta1.ScopeRef{Name: "s"},
				IpRange: cloudcontrolv1beta1.IpRangeRef{
					Name: "foo",
				},
				Instance: cloudcontrolv1beta1.NfsInstanceInfo{
					Aws: &cloudcontrolv1beta1.NfsInstanceAws{},
				},
			},
		}

		err = infra.KCP().Client().Create(infra.Ctx(), obj)
		Expect(err).NotTo(HaveOccurred())

		_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
	})

	// GCP ============================================

	It("Scenario: KCP NfsInstance GCP without IpRange can not be created", func() {
		name := uuid.NewString()
		var err error
		obj := &cloudcontrolv1beta1.NfsInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: infra.KCP().Namespace(),
			},
			Spec: cloudcontrolv1beta1.NfsInstanceSpec{
				Scope: cloudcontrolv1beta1.ScopeRef{Name: "s"},
				Instance: cloudcontrolv1beta1.NfsInstanceInfo{
					Gcp: &cloudcontrolv1beta1.NfsInstanceGcp{
						Location:      "us-east-1",
						ConnectMode:   cloudcontrolv1beta1.PRIVATE_SERVICE_ACCESS,
						FileShareName: "vol1",
						Tier:          "BASIC_SSD",
					},
				},
			},
		}

		err = infra.KCP().Client().Create(infra.Ctx(), obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("IpRange can not be specified for openstack, and is mandatory for gcp and aws"))

		_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
	})

	It("Scenario: KCP NfsInstance GCP with IpRange can be created", func() {
		name := uuid.NewString()
		var err error
		obj := &cloudcontrolv1beta1.NfsInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: infra.KCP().Namespace(),
			},
			Spec: cloudcontrolv1beta1.NfsInstanceSpec{
				Scope: cloudcontrolv1beta1.ScopeRef{Name: "s"},
				IpRange: cloudcontrolv1beta1.IpRangeRef{
					Name: "foo",
				},
				Instance: cloudcontrolv1beta1.NfsInstanceInfo{
					Gcp: &cloudcontrolv1beta1.NfsInstanceGcp{
						Location:      "us-east-1",
						ConnectMode:   cloudcontrolv1beta1.PRIVATE_SERVICE_ACCESS,
						FileShareName: "vol1",
						Tier:          "BASIC_SSD",
					},
				},
			},
		}

		err = infra.KCP().Client().Create(infra.Ctx(), obj)
		Expect(err).NotTo(HaveOccurred())

		_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
	})
})
