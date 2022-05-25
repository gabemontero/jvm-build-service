package dependencybuild

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-appstudio/jvm-build-service/pkg/reconciler/artifactbuildrequest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"

	"github.com/redhat-appstudio/jvm-build-service/pkg/apis/jvmbuildservice/v1alpha1"

	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

func setupClientAndReconciler(objs ...runtimeclient.Object) (runtimeclient.Client, *ReconcileDependencyBuild) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	_ = pipelinev1beta1.AddToScheme(scheme)
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	reconciler := &ReconcileDependencyBuild{client: client, scheme: scheme}
	return client, reconciler
}

func TestSetup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dependency Build Suite")
}

var _ = It("Test reconcile new DependencyBuild", func() {
	db := v1alpha1.DependencyBuild{}
	db.Namespace = metav1.NamespaceDefault
	db.Name = "test"
	db.Status.State = v1alpha1.DependencyBuildStateNew
	db.Spec.SCMURL = "some-url"
	db.Spec.Tag = "some-tag"
	db.Spec.Path = "some-path"
	db.Labels = map[string]string{artifactbuildrequest.DependencyBuildIdLabel: hashToString(db.Spec.SCMURL + db.Spec.Tag + db.Spec.Path)}

	ctx := context.TODO()
	client, reconciler := setupClientAndReconciler(&db)

	Expect(reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Namespace: db.Namespace, Name: db.Name}}))

	Expect(client.Get(ctx, types.NamespacedName{
		Namespace: metav1.NamespaceDefault,
		Name:      "test",
	}, &db))
	Expect(db.Status.State).Should(Equal(v1alpha1.DependencyBuildStateBuilding))

	prList := &pipelinev1beta1.PipelineRunList{}
	Expect(client.List(ctx, prList))

	Expect(len(prList.Items)).Should(Equal(1))
	for _, pr := range prList.Items {
		Expect(pr.Labels[artifactbuildrequest.DependencyBuildIdLabel]).Should(Equal(db.Labels[artifactbuildrequest.DependencyBuildIdLabel]))
		for _, or := range pr.OwnerReferences {
			if or.Kind != db.Kind || or.Name != db.Name {
				Expect(or.Kind).Should(Equal(db.Kind))
				Expect(or.Name).Should(Equal(db.Name))
			}
		}
		Expect(len(pr.Spec.Params)).Should(Equal(3))
		for _, param := range pr.Spec.Params {
			switch param.Name {
			case PipelineScmTag:
				Expect(param.Value.StringVal).Should(Equal("some-tag"))
			case PipelinePath:
				Expect(param.Value.StringVal).Should(Equal("some-path"))
			case PipelineScmUrl:
				Expect(param.Value.StringVal).Should(Equal("some-url"))
			}
		}
	}
})
