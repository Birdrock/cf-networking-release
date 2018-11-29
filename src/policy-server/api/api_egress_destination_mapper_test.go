package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"
	"policy-server/api"
	"policy-server/api/fakes"
	"policy-server/store"

	"errors"

	"code.cloudfoundry.org/cf-networking-helpers/marshal"
)

var _ = Describe("ApiEgressDestinationMapper", func() {
	var (
		mapper        *api.EgressDestinationMapper
		fakeValidator *fakes.EgressDestinationsValidator
	)

	BeforeEach(func() {
		fakeValidator = &fakes.EgressDestinationsValidator{}
		fakeValidator.ValidateEgressDestinationsReturns(nil)

		mapper = &api.EgressDestinationMapper{
			Marshaler:        marshal.MarshalFunc(json.Marshal),
			PayloadValidator: fakeValidator,
		}
	})

	Describe("AsBytes", func() {
		var egressDestinations []store.EgressDestination
		BeforeEach(func() {
			egressDestinations = []store.EgressDestination{
				{
					GUID:     "1",
					Name:     " ",
					Protocol: "tcp",
					Ports: []store.Ports{{
						Start: 8080,
						End:   8081,
					}},
					IPRanges: []store.IPRange{{
						Start: "1.2.3.4",
						End:   "1.2.3.5",
					}},
				},
				{
					GUID:        "2",
					Description: " ",
					Protocol:    "icmp",
					IPRanges: []store.IPRange{{
						Start: "1.2.3.7",
						End:   "1.2.3.8",
					}},
					ICMPType: 1,
					ICMPCode: 6,
				},
				{
					GUID:     "3",
					Protocol: "udp",
					IPRanges: []store.IPRange{{
						Start: "1.2.3.7",
						End:   "1.2.3.8",
					}},
				},
			}
		})

		It("marshals to json", func() {
			payload, err := mapper.AsBytes(egressDestinations)
			Expect(err).NotTo(HaveOccurred())
			Expect(payload).To(MatchJSON(
				[]byte(`{
					"total_destinations": 3,
					"destinations": [
						{
							"id": "1",
							"name": " ",
							"protocol": "tcp",
							"ports": [{ "start": 8080, "end": 8081 }],
							"ips": [{ "start": "1.2.3.4", "end": "1.2.3.5" }]
						},
						{
							"id": "2",
							"description": " ",
							"protocol": "icmp",
							"ips": [{ "start": "1.2.3.7", "end": "1.2.3.8" }],
							"icmp_type": 1,
							"icmp_code": 6
						},
 						{
							"id": "3",
							"protocol": "udp",
							"ips": [{ "start": "1.2.3.7", "end": "1.2.3.8" }]
						}
					]
				}`)))
		})
	})

	Describe("AsEgressDestinations", func() {
		var expectedOutputBytes []byte

		BeforeEach(func() {
			expectedOutputBytes = []byte(`{
					"total_destinations": 3,
					"destinations": [
						{
							"id": "1",
							"name": "my service",
							"protocol": "tcp",
							"ports": [{ "start": 8080, "end": 8081 }],
							"ips": [{ "start": "1.2.3.4", "end": "1.2.3.5" }],
							"app_lifecycle": "all"
						},
						{
							"id": "2",
							"description": "this is where my apps go",
							"protocol": "icmp",
							"ips": [{ "start": "1.2.3.7", "end": "1.2.3.8" }],
							"icmp_type": 1,
							"icmp_code": 6,
							"app_lifecycle": "all"
						},
						{
							"id": "3",
							"description": "regression test: icmp without type and code",
							"protocol": "icmp",
							"ips": [{ "start": "1.2.3.7", "end": "1.2.3.8" }],
							"app_lifecycle": "all"
						},
						{
							"id": "4",
							"protocol": "udp",
							"ips": [{ "start": "1.2.3.7", "end": "1.2.3.8" }],
							"app_lifecycle": "all"
						}
					]
				}`)
		})

		It("unmarshals egress destinations from json", func() {
			payload, err := mapper.AsEgressDestinations(expectedOutputBytes)
			Expect(err).NotTo(HaveOccurred())
			Expect(payload).To(Equal(
				[]store.EgressDestination{
					{
						GUID:     "1",
						Name:     "my service",
						Protocol: "tcp",
						Ports: []store.Ports{{
							Start: 8080,
							End:   8081,
						}},
						IPRanges: []store.IPRange{{
							Start: "1.2.3.4",
							End:   "1.2.3.5",
						}},
					},
					{
						GUID:        "2",
						Description: "this is where my apps go",
						Protocol:    "icmp",
						Ports:       []store.Ports{},
						IPRanges: []store.IPRange{{
							Start: "1.2.3.7",
							End:   "1.2.3.8",
						}},
						ICMPType: 1,
						ICMPCode: 6,
					},
					{
						GUID:        "3",
						Description: "regression test: icmp without type and code",
						Protocol:    "icmp",
						Ports:       []store.Ports{},
						IPRanges: []store.IPRange{{
							Start: "1.2.3.7",
							End:   "1.2.3.8",
						}},
						ICMPType: -1,
						ICMPCode: -1,
					},
					{
						GUID:     "4",
						Protocol: "udp",
						Ports:    []store.Ports{},
						IPRanges: []store.IPRange{{
							Start: "1.2.3.7",
							End:   "1.2.3.8",
						}},
					},
				}),
			)
		})

		Context("when there is a json unmarshalling error", func() {
			It("returns an error", func() {
				_, err := mapper.AsEgressDestinations([]byte("%%%"))
				Expect(err).To(MatchError("unmarshal json: invalid character '%' looking for beginning of value"))
			})
		})

		Context("when there is a validation error", func() {
			BeforeEach(func() {
				fakeValidator.ValidateEgressDestinationsReturns(errors.New("banana"))
			})

			It("returns an error", func() {
				_, err := mapper.AsEgressDestinations(expectedOutputBytes)
				Expect(err).To(MatchError("validate destinations: banana"))
			})
		})
	})
})
