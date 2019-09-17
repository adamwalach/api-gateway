package processing

import (
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

func prepareAccessRule(api *gatewayv2alpha1.Gate, ar *rulev1alpha1.Rule, rule gatewayv2alpha1.Rule, accessStrategies []*rulev1alpha1.Authenticator) *rulev1alpha1.Rule {
	ar.ObjectMeta.OwnerReferences = []k8sMeta.OwnerReference{generateOwnerRef(api)}
	ar.ObjectMeta.Name = fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	ar.ObjectMeta.Namespace = api.ObjectMeta.Namespace

	return builders.AccessRule().From(ar).
		Spec(builders.AccessRuleSpec().From(generateAccessRuleSpec(api, rule, accessStrategies))).
		Get()
}

func generateAccessRule(api *gatewayv2alpha1.Gate, rule gatewayv2alpha1.Rule, accessStrategies []*rulev1alpha1.Authenticator) *rulev1alpha1.Rule {
	name := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	namespace := api.ObjectMeta.Namespace
	ownerRef := generateOwnerRef(api)

	return builders.AccessRule().
		Name(name).
		Namespace(namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(builders.AccessRuleSpec().From(generateAccessRuleSpec(api, rule, accessStrategies))).
		Get()
}

func generateAccessRuleSpec(api *gatewayv2alpha1.Gate, rule gatewayv2alpha1.Rule, accessStrategies []*rulev1alpha1.Authenticator) *rulev1alpha1.RuleSpec {
	return builders.AccessRuleSpec().
		Upstream(builders.Upstream().
			URL(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", *api.Spec.Service.Name, api.ObjectMeta.Namespace, int(*api.Spec.Service.Port)))).
		Match(builders.Match().
			URL(fmt.Sprintf("<http|https>://%s<%s>", *api.Spec.Service.Host, rule.Path)).
			Methods(rule.Methods)).
		Authorizer(builders.Authorizer().Handler(builders.Handler().
			Name("allow"))).
		Authenticators(builders.Authenticators().From(accessStrategies)).
		Mutators(builders.Mutators().From(rule.Mutators)).Get()
}

func generateVirtualService(api *gatewayv2alpha1.Gate, destinationHost string, destinationPort uint32, path string) *networkingv1alpha3.VirtualService {
	virtualServiceName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	ownerRef := generateOwnerRef(api)
	return builders.VirtualService().
		Name(virtualServiceName).
		Namespace(api.ObjectMeta.Namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(
			builders.VirtualServiceSpec().
				Host(*api.Spec.Service.Host).
				Gateway(*api.Spec.Gateway).
				HTTP(
					builders.MatchRequest().URI().Regex(path),
					builders.RouteDestination().Host(destinationHost).Port(destinationPort))).
		Get()
}

func isSecured(rule gatewayv2alpha1.Rule) bool {
	if len(rule.Scopes) > 0 || len(rule.Mutators) > 0 {
		return true
	}
	return false
}

func generateOwnerRef(api *gatewayv2alpha1.Gate) k8sMeta.OwnerReference {
	return *builders.OwnerReference().
		Name(api.ObjectMeta.Name).
		APIVersion(api.TypeMeta.APIVersion).
		Kind(api.TypeMeta.Kind).
		UID(api.ObjectMeta.UID).
		Controller(true).
		Get()
}

func generateObjectMeta(api *gatewayv2alpha1.Gate) k8sMeta.ObjectMeta {
	ownerRef := generateOwnerRef(api)
	return *builders.ObjectMeta().
		Name(fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)).
		Namespace(api.ObjectMeta.Namespace).
		OwnerReference(builders.OwnerReference().From(&ownerRef)).
		Get()
}

func prepareVirtualService(api *gatewayv2alpha1.Gate, vs *networkingv1alpha3.VirtualService, destinationHost string, destinationPort uint32, path string) *networkingv1alpha3.VirtualService {
	virtualServiceName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	ownerRef := generateOwnerRef(api)

	return builders.VirtualService().From(vs).
		Name(virtualServiceName).
		Namespace(api.ObjectMeta.Namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Spec(
			builders.VirtualServiceSpec().
				Host(*api.Spec.Service.Host).
				Gateway(*api.Spec.Gateway).
				HTTP(
					builders.MatchRequest().URI().Regex(path),
					builders.RouteDestination().Host(destinationHost).Port(destinationPort))).
		Get()
}